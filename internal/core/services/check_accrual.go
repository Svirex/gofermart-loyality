package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Svirex/gofermart-loyality/internal/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type CheckAccrualService struct {
	dbpool               *pgxpool.Pool
	stopCh               chan struct{}
	orderNumsCh          chan string
	queueSize            int
	accrualAddr          string
	errorCh              chan error
	logger               common.Logger
	accrualResponseCh    chan *AccrualResponse
	pauseBetweenRequests time.Duration
	generatorsWG         sync.WaitGroup
	checkerEndCh         chan struct{}
	// dbWriterEndCh        chan struct{}
	errorLogEndCh        chan struct{}
	dbLoaderEndCh        chan struct{}
	currentGeneratorsRun atomic.Int32
	maxRunnedGenerators  int32
	dbLoaderPause        time.Duration
}

func NewCheckAccrualService(dbpool *pgxpool.Pool,
	logger common.Logger,
	queueSize int,
	accrualAddr string,
	pauseBetweenRequests time.Duration,
	maxRunnedGenerators int32,
	dbLoaderPause time.Duration,
) (*CheckAccrualService, error) {
	return &CheckAccrualService{
		dbpool:               dbpool,
		stopCh:               make(chan struct{}),
		orderNumsCh:          make(chan string, queueSize),
		queueSize:            queueSize,
		accrualAddr:          accrualAddr,
		errorCh:              make(chan error, queueSize),
		logger:               logger,
		accrualResponseCh:    make(chan *AccrualResponse, queueSize),
		pauseBetweenRequests: pauseBetweenRequests,
		checkerEndCh:         make(chan struct{}),
		// dbWriterEndCh:        make(chan struct{}),
		errorLogEndCh:       make(chan struct{}),
		dbLoaderEndCh:       make(chan struct{}),
		maxRunnedGenerators: maxRunnedGenerators,
		dbLoaderPause:       dbLoaderPause,
	}, nil
}

// сервис старутет, мы должны из базы вычитать N номер заказов в неконечном статусе и положить их в очередь
//

func (service *CheckAccrualService) Start() {
	service.logger.Debug("START CHECK ACCRUAL SERVICE")
	go service.dbLoader()
	go service.checker()
	// go service.dbWriter()
	go service.errorLog()
}

func (service *CheckAccrualService) dbLoader() {
	service.logger.Debug("START DB LOADER")
	for {
		select {
		case <-service.stopCh:
			close(service.dbLoaderEndCh)
			service.logger.Debugln("CLOSE CHANNEL dbLoaderEndCh")
			return
		default:
			service.logger.Debug("DB LOADER currentGeneratorsRun: ", service.currentGeneratorsRun.Load(), ", maxRunnedGenerators: ", service.maxRunnedGenerators)
			if service.currentGeneratorsRun.Load() < service.maxRunnedGenerators {
				row, _ := service.dbpool.Query(context.Background(), "SELECT order_num FROM orders WHERE status='NEW' OR status='PROCESSING';")
				if err := row.Err(); err != nil {
					service.errorCh <- fmt.Errorf("check accrual service, db loader, select orders num: %w", err)
					break
				}
				orderNums := make([]string, 0)
				for row.Next() {
					if err := row.Err(); err != nil {
						service.errorCh <- fmt.Errorf("check accrual service, db loader, row.Next: %w", err)
						break
					}
					var orderNum string
					if err := row.Scan(&orderNum); err != nil {
						service.errorCh <- fmt.Errorf("check accrual service, db loader, scan order num: %w", err)
						break
					}
					orderNums = append(orderNums, orderNum)
				}
				service.logger.Debug("DB LOADER orderNums: ", orderNums)
				for i := range orderNums {
					service.Process(orderNums[i])
				}
			}
			time.Sleep(service.dbLoaderPause)
		}
	}

}

func (service *CheckAccrualService) startGenerator() {
	service.generatorsWG.Add(1)
	service.currentGeneratorsRun.Add(1)
}

func (service *CheckAccrualService) endGenerator() {
	service.generatorsWG.Done()
	service.currentGeneratorsRun.Add(-1)
}

func (service *CheckAccrualService) Process(orderNum string) {
	if service.currentGeneratorsRun.Load() < service.maxRunnedGenerators {
		select {
		case <-service.stopCh:
			return
		default:
			service.startGenerator()
			go service.generator(orderNum)
		}
	}
}

func (service *CheckAccrualService) generator(orderNum string) {
	select {
	case <-service.stopCh:
		service.endGenerator()
		return
	default:
		service.orderNumsCh <- orderNum
		service.endGenerator()
		return
	}
}

var perMinuteRegexp = regexp.MustCompile(`\d+`)

// вычитывает из orderNumsCh номер заказа и делает GET запрос в сервис проверки
// и кладёт объект в канал, который обрабатывает заказ
func (service *CheckAccrualService) checker() {
	for {
		select {
		case <-service.stopCh:
			close(service.checkerEndCh)
			service.logger.Debugln("CLOSE CHANNEL checkerEndCh")
			return
		case orderNum := <-service.orderNumsCh:
			service.logger.Debugln("ORDER_NUM", orderNum)
			request, err := http.NewRequest(http.MethodGet, service.accrualAddr+"/api/orders/"+orderNum, http.NoBody)
			if err != nil {
				service.errorCh <- fmt.Errorf("check accrual service, checker, new request: %w", err)
				service.orderNumsCh <- orderNum // чтобы не упустить из обработки orderNum
				break
			}
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				service.errorCh <- fmt.Errorf("check accrual service, checker, do request: %w", err)
				service.orderNumsCh <- orderNum // чтобы не упустить из обработки orderNum
				break
			}
			if response.StatusCode == http.StatusOK {
				data, err := getAccrualResponse(response)
				if err != nil {
					service.errorCh <- fmt.Errorf("check accrual service, checker, response status ok, get accrual response: %w", err)
					service.orderNumsCh <- orderNum // чтобы не упустить из обработки orderNum
					break
				}
				response.Body.Close()
				// service.accrualResponseCh <- data
				service.writeData(data)

			} else if response.StatusCode == http.StatusNoContent {
				service.orderNumsCh <- orderNum // чтобы не упустить из обработки orderNum
			} else if response.StatusCode == http.StatusTooManyRequests {
				// вычитать данные из заголовка Retry-After, сохранить как количество секунд
				// получить Н из "No more than N requests per minute allowed", чтобы определить паузу между запросам
				var retryAfter int
				if retryAfterString := response.Header.Get("Retry-After"); len(retryAfterString) != 0 {
					v, err := strconv.Atoi(retryAfterString)
					if err != nil {
						service.errorCh <- fmt.Errorf("check accrual service, checker, retry-after header atoi: %w", err)
						service.orderNumsCh <- orderNum // чтобы не упустить из обработки orderNum
						break
					}
					retryAfter = v
					service.logger.Debugln("retry after ", retryAfter)
				}
				body, err := io.ReadAll(response.Body)
				if err != nil {
					service.errorCh <- fmt.Errorf("check accrual service, checker, 429, read body: %w", err)
					service.orderNumsCh <- orderNum // чтобы не упустить из обработки orderNum
					break
				}
				response.Body.Close()
				requestPerMinuteStr := perMinuteRegexp.Find(body)
				service.logger.Debugln("PER MINUTE REGEXP FIND ", string(requestPerMinuteStr))
				if len(requestPerMinuteStr) != 0 {
					v, err := strconv.Atoi(string(requestPerMinuteStr))
					if err != nil {
						service.errorCh <- fmt.Errorf("check accrual service, checker, 429, requestPerMinuteStr atoi: %w", err)
						service.orderNumsCh <- orderNum // чтобы не упустить из обработки orderNum
						break
					}
					service.logger.Debugln("REQUEST PER MINUTE ", v)
					service.pauseBetweenRequests = time.Duration(60/v) * time.Second
					service.logger.Debugln("NEW PAUSE BETWEEN REQUESTS ", service.pauseBetweenRequests)
				}
				time.Sleep(time.Duration(retryAfter) - service.pauseBetweenRequests)
			}
			time.Sleep(service.pauseBetweenRequests)
		}
	}

}

func (service *CheckAccrualService) writeData(ar *AccrualResponse) {
	service.logger.Debugln("service.accrualResponseCh", ar, decimal.Decimal(ar.Accrual).String())
	switch ar.Status {
	case Registered:
		service.orderNumsCh <- ar.OrderNum
	case Invalid:
		service.logger.Debugln("INVALID ORDER_NUM", ar.OrderNum, ar)
		err := service.writeInvalid(ar.OrderNum)
		if err != nil {
			service.errorCh <- fmt.Errorf("dbWriter, invalid: %w", err)
			// service.accrualResponseCh <- ar // чтобы попробовать записать ещё раз
		}
	case Processing:
		err := service.writeProcessing(ar.OrderNum)
		if err != nil {
			service.errorCh <- fmt.Errorf("dbWriter, processing: %w", err)
			// service.accrualResponseCh <- ar // чтобы попробовать записать ещё раз
		}
		service.orderNumsCh <- ar.OrderNum
	case Processed:
		err := service.writeProcessed(ar)
		if err != nil {
			service.errorCh <- fmt.Errorf("dbWriter, procedd: %w", err)
			// service.accrualResponseCh <- ar // чтобы попробовать записать ещё раз
		}
	}
}

// func (service *CheckAccrualService) dbWriter() {
// 	for {
// 		select {
// 		case <-service.stopCh:
// 			close(service.dbWriterEndCh)
// 			service.logger.Debugln("CLOSE CHANNEL dbWriterEndCh")
// 			return
// 		case ar := <-service.accrualResponseCh:
// 			service.logger.Debugln("service.accrualResponseCh", ar, decimal.Decimal(ar.Accrual).String())
// 			switch ar.Status {
// 			case Registered:
// 				service.orderNumsCh <- ar.OrderNum
// 			case Invalid:
// 				service.logger.Debugln("INVALID ORDER_NUM", ar.OrderNum, ar)
// 				err := service.writeInvalid(ar.OrderNum)
// 				if err != nil {
// 					service.errorCh <- fmt.Errorf("dbWriter, invalid: %w", err)
// 					service.accrualResponseCh <- ar // чтобы попробовать записать ещё раз
// 				}
// 			case Processing:
// 				err := service.writeProcessing(ar.OrderNum)
// 				if err != nil {
// 					service.errorCh <- fmt.Errorf("dbWriter, processing: %w", err)
// 					service.accrualResponseCh <- ar // чтобы попробовать записать ещё раз
// 				}
// 				service.orderNumsCh <- ar.OrderNum
// 			case Processed:
// 				err := service.writeProcessed(ar)
// 				if err != nil {
// 					service.errorCh <- fmt.Errorf("dbWriter, procedd: %w", err)
// 					service.accrualResponseCh <- ar // чтобы попробовать записать ещё раз
// 				}
// 			}
// 		}
// 	}
// }

func (service *CheckAccrualService) writeInvalid(orderNum string) error {
	_, err := service.dbpool.Exec(context.Background(), "UPDATE orders SET status='INVALID' WHERE order_num=$1;", orderNum)
	if err != nil {
		return fmt.Errorf("write invalid: %w", err)
	}
	return nil
}

func (service *CheckAccrualService) writeProcessing(orderNum string) error {
	_, err := service.dbpool.Exec(context.Background(), "UPDATE orders SET status='PROCESSING' WHERE order_num=$1;", orderNum)
	if err != nil {
		return fmt.Errorf("write processing: %w", err)
	}
	return nil
}

func (service *CheckAccrualService) writeProcessed(ar *AccrualResponse) error {
	var uid int64
	err := service.dbpool.QueryRow(context.Background(), "UPDATE orders SET status='PROCESSED' WHERE order_num=$1 AND status!='PROCESSED' RETURNING uid;", ar.OrderNum).Scan(&uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("write processed, update status: %w", err)
	}
	if decimal.Decimal(ar.Accrual).IsZero() {
		return nil
	}

	service.logger.Debug("SELECTED UID ", uid)

	_, err = service.dbpool.Exec(context.Background(), "UPDATE balance SET current=current+$1 WHERE uid=$2;", decimal.Decimal(ar.Accrual).String(), uid)
	if err != nil {
		return fmt.Errorf("write processed, update: %w", err)
	}
	return nil
}

func (service *CheckAccrualService) errorLog() {
	for {
		select {
		case <-service.stopCh:
			close(service.errorLogEndCh)
			service.logger.Debugln("CLOSE CHANNEL errorLogEndCh")
			return
		case err := <-service.errorCh:
			service.logger.Errorf("CheckAccrualService: %v", err)
		}
	}
}

func (service *CheckAccrualService) Shutdown() {
	close(service.stopCh)
	service.logger.Debugln("LEN orderNumsCh", len(service.orderNumsCh))
	for len(service.orderNumsCh) > 0 {
		service.logger.Debugln("LEN orderNumsCh read ", len(service.orderNumsCh))
		<-service.orderNumsCh
	}
	service.generatorsWG.Wait()
	<-service.errorLogEndCh
	<-service.checkerEndCh
	// <-service.dbWriterEndCh
	<-service.dbLoaderEndCh
	close(service.orderNumsCh)
	service.logger.Debugln("CLOSE CHANNEL orderNumsCh")
	close(service.errorCh)
	service.logger.Debugln("CLOSE CHANNEL errorCh")
	close(service.accrualResponseCh)
	service.logger.Debugln("CLOSE CHANNEL accrualResponseCh")
}

func getAccrualResponse(response *http.Response) (*AccrualResponse, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("getAccrualResponse, read body: %w", err)
	}
	ar := &AccrualResponse{}
	if err = json.Unmarshal(body, &ar); err != nil {
		return nil, fmt.Errorf("getAccrualResponse, unmarshal body: %w, body: %v", err, string(body))
	}
	return ar, nil
}

type Amount decimal.Decimal

func (a *Amount) UnmarshalJSON(b []byte) error {
	var s float64
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("unmarshal amount: %w", err)
	}
	v := decimal.NewFromFloat(s)
	*a = Amount(v)
	return nil
}

func (a Amount) MarshalJSON() ([]byte, error) {
	s := decimal.Decimal(a).String()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("marshal amount: %w", err)
	}
	return json.Marshal(v)
}

type Status int

func (status *Status) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("unmarshal status: %w", err)
	}
	switch s {
	case "REGISTERED":
		*status = Registered
	case "INVALID":
		*status = Invalid
	case "PROCESSING":
		*status = Processing
	case "PROCESSED":
		*status = Processed
	default:
		return fmt.Errorf("unmarshal status, unknown status")
	}
	return nil
}

func (status Status) MarshalJSON() ([]byte, error) {
	var s string
	switch status {
	case Registered:
		s = "REGISTERED"
	case Invalid:
		s = "INVALID"
	case Processing:
		s = "PROCESSING"
	case Processed:
		s = "PROCESSED"
	}
	return json.Marshal(s)
}

const (
	Registered Status = iota
	Invalid
	Processing
	Processed
)

type AccrualResponse struct {
	OrderNum string `json:"order"`
	Status   Status `json:"status"`
	Accrual  Amount `json:"accrual"`
}
