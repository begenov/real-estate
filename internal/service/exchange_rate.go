package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/begenov/real-estate/internal/async"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/repository/postgres"
	"net/http"
	"time"
)

type exchangeRateResponse struct {
	Success bool `json:"success"`
	Quotes  map[string]struct {
		EndRate float64 `json:"end_rate"`
	} `json:"quotes"`
}

type IExchangeRateService interface {
	StartUpdater(ctx context.Context)
	GetRates(ctx context.Context) (map[string]float64, error)
}

type ExchangeRateService struct {
	repo   postgres.IExchangeRateRepo
	apiKey string
	sf     async.Singleflight
}

func NewExchangeRateService(repo postgres.IExchangeRateRepo, apiKey string) IExchangeRateService {
	return &ExchangeRateService{repo: repo, apiKey: apiKey}
}

func (s *ExchangeRateService) StartUpdater(ctx context.Context) {
	//if err := s.updateRates(ctx); err != nil {
	//	logger.Error("Failed to get rates on startup: ", err)
	//}

	ticker := time.NewTicker(time.Hour * 24)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.updateRates(ctx); err != nil {
				logger.Error("Failed to get rates: ", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *ExchangeRateService) updateRates(ctx context.Context) error {
	usd, eur, err := s.getRatesGBPtoUSDandEUR(ctx)
	if err != nil {
		return err
	}
	return s.repo.UpdateExchangeRates(ctx, map[string]float64{
		"USD": usd,
		"EUR": eur,
	})
}

func (s *ExchangeRateService) GetRates(ctx context.Context) (map[string]float64, error) {
	val, err, _ := s.sf.Do("exchange_rates", func() (interface{}, error) {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return s.repo.GetExchangeRates(ctx)
	})
	if err != nil {
		return nil, err
	}
	rates, ok := val.(map[string]float64)
	if !ok {
		return nil, errors.New("unexpected exchange rate response type")
	}
	return rates, nil
}

func (s *ExchangeRateService) getRatesGBPtoUSDandEUR(ctx context.Context) (usdRate, eurRate float64, err error) {
	if s.apiKey == "" {
		return usdRate, eurRate, errors.New("api key is empty")
	}

	url := fmt.Sprintf("https://api.exchangerate.host/change?access_key=%s&source=GBP&currencies=USD,EUR&format=1", s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var data exchangeRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, err
	}

	usdRate = data.Quotes["GBPUSD"].EndRate
	eurRate = data.Quotes["GBPEUR"].EndRate
	return usdRate, eurRate, nil
}
