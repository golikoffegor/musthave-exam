package main

import (
	"context"
	"musthave-exam/internal/app"
	"musthave-exam/internal/handler"
	"musthave-exam/internal/process"
	"musthave-exam/internal/repository"
	"musthave-exam/internal/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	ap, err := app.NewApp()
	if err != nil {
		ap.Logger.Fatal(err)
	}
	defer ap.DB.Close()

	go handleSignals(ap.CncF, ap.Logger)

	//set repo
	repo := repository.NewRepository(ap.DB, ap.Logger)
	//set handlers
	vhandler := handler.NewHandler(repo, ap.Logger, ap.Flags)

	ap.Logger.WithField("Flags", *ap.Flags).Info("App init")
	proc := process.NewOrderProcess(repo, ap.Logger, (*ap.Flags).RecalcAddress)
	go proc.StartTransactionProcessing(ap.Ctx)

	go func() {
		ap.Logger.Info("Start ListenAndServe")
		ap.Logger.Fatal(http.ListenAndServe(ap.Flags.Endpoint, router.InitRouter(*vhandler)))
	}()

	<-ap.Ctx.Done()
	ap.Logger.Info("Server stopped gracefully")
}

func handleSignals(cancel context.CancelFunc, logger *logrus.Logger) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Ожидаем сигнал
	sig := <-sigs
	logger.Info("Received signal: ", sig)
	cancel()
}

// type OrderResponse struct {
// 	Order   string   `json:"order"`
// 	Status  string   `json:"status"`
// 	Accrual *float64 `json:"accrual,omitempty"`
// }

// func checkOrderStatus(ctx context.Context, orderNumber string) (*OrderResponse, error) {
// 	url := fmt.Sprintf("http://localhost:8010/api/orders/%s", orderNumber)
// 	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode == http.StatusNoContent {
// 		return nil, nil
// 	}

// 	var orderResponse OrderResponse
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	json.Unmarshal(body, &orderResponse)
// 	fmt.Printf("orderResponse: %v\n", orderResponse)

// 	return &orderResponse, nil
// }

// func monitorOrderStatus(ctx context.Context) {
// 	orderNumber := "7785644118"
// 	interval := 1 * time.Second
// 	log.Printf("Starting to monitor order status for order number: %s", orderNumber)

// 	ticker := time.NewTicker(interval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			orderResponse, err := checkOrderStatus(ctx, orderNumber)
// 			if err != nil {
// 				log.Printf("Failed to check order status: %v", err)
// 				continue
// 			}

// 			if orderResponse != nil {
// 				log.Printf("Order status: %s", orderResponse.Status)
// 				if orderResponse.Status == "PROCESSED" || orderResponse.Status == "INVALID" {
// 					log.Printf("Final status received for order %s: %s", orderNumber, orderResponse.Status)
// 					return
// 				}
// 			} else {
// 				log.Println("Order not found")
// 			}
// 		case <-ctx.Done():
// 			log.Println("Monitoring cancelled")
// 			return
// 		}
// 	}
// }
