package service

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testtask/model"
	"time"
)

type ApiResponse struct {
	Raw     map[string]map[string]innerDataRaw     `json:"RAW"`
	Display map[string]map[string]innerDataDisplay `json:"DISPLAY"`
}

type innerDataRaw struct {
	Price           float64 `json:"PRICE"`
	Volume24Hour    float64 `json:"VOLUME24HOUR"`
	Volume24Hourto  float64 `json:"VOLUME24HOURTO"`
	Open24Hour      float64 `json:"OPEN24HOUR"`
	High24Hour      float64 `json:"HIGH24HOUR"`
	Low24Hour       float64 `json:"LOW24HOUR"`
	Change24Hour    float64 `json:"CHANGE24HOUR"`
	Changepct24Hour float64 `json:"CHANGEPCT24HOUR"`
	Supply          float64 `json:"SUPPLY"`
	Mktcap          float64 `json:"MKTCAP"`
}

type innerDataDisplay struct {
	Price           string `json:"PRICE"`
	Volume24Hour    string `json:"VOLUME24HOUR"`
	Volume24Hourto  string `json:"VOLUME24HOURTO"`
	Open24Hour      string `json:"OPEN24HOUR"`
	High24Hour      string `json:"HIGH24HOUR"`
	Low24Hour       string `json:"LOW24HOUR"`
	Change24Hour    string `json:"CHANGE24HOUR"`
	Changepct24Hour string `json:"CHANGEPCT24HOUR"`
	Supply          string `json:"SUPPLY"`
	Mktcap          string `json:"MKTCAP"`
}

type ApiQuery struct {
	Fsyms []string
	Tsyms []string
}

type PriceManager struct {
	priceRepository *model.PriceRepository
	data            string
	store           chan []byte
	cache           map[string][]byte
	updated         *time.Time
}

func CreatePriceService(pr *model.PriceRepository) *PriceManager {
	return &PriceManager{
		priceRepository: pr,
		data:            "",
		store:           make(chan []byte),
		cache:           map[string][]byte{},
	}
}

func (m *PriceManager) loadFromCache(query *ApiQuery) ([]byte, error) {
	if m.updated == nil {
		return nil, fmt.Errorf("cache is absent")
	}

	if time.Since(*m.updated).Minutes() > 1 {
		return nil, fmt.Errorf("cache is too old")
	}

	fsyms := string(strings.Join(query.Fsyms[:], ","))
	tsyms := string(strings.Join(query.Tsyms[:], ","))
	bytehash := sha1.Sum([]byte(fsyms + tsyms))
	data, ok := m.cache[string(bytehash[:])]

	if ok {
		return data, nil
	}

	return nil, fmt.Errorf("cache is absent")
}

func (m *PriceManager) saveToCache(query *ApiQuery, data []byte) {
	fsyms := string(strings.Join(query.Fsyms[:], ","))
	tsyms := string(strings.Join(query.Tsyms[:], ","))
	bytehash := sha1.Sum([]byte(fsyms + tsyms))
	m.cache[string(bytehash[:])] = data
	t := time.Now()
	m.updated = &t
}

func (m *PriceManager) GetPrices(fsyms []string, tsyms []string) ([]byte, error) {
	query := &ApiQuery{
		Fsyms: fsyms,
		Tsyms: tsyms,
	}

	result, err := m.loadFromCache(query)

	if err != nil {
		result, err = loadApiData(query)

		if err != nil {
			result, err = loadLocalApiData(m.priceRepository, query)
		}

		if err != nil {
			return nil, fmt.Errorf("resource not available")
		}
		m.saveToCache(query, result)
	}

	return result, nil
}

// Broadcast should be run as goroutine only because of blocking calls
func (m *PriceManager) Broadcast(client chan []byte, fsyms []string, tsyms []string) error {
	wg := sync.WaitGroup{}
	closed := false
	dataUnavailable := false

	for {
		wg.Add(1)
		go func() {
			defer wg.Done()
			query := &ApiQuery{
				Fsyms: fsyms,
				Tsyms: tsyms,
			}

			result, err := m.loadFromCache(query)

			if err != nil {
				result, err = loadApiData(query)
				if err != nil {
					result, err = loadLocalApiData(m.priceRepository, query)
				}

				if err != nil {
					closed = true
					dataUnavailable = true
					return
				}

				storeApiData(m.priceRepository, query, result)
			}

			closed = SafeSend(client, result)
		}()
		wg.Wait()
		// Gracefully exits if client channel is closed
		if closed {
			break
		}
		// sleep a bit to lower the load on the server
		time.Sleep(time.Second * 5)
	}

	if dataUnavailable {
		return fmt.Errorf("resource not available")
	}

	return nil
}

func loadLocalApiData(r *model.PriceRepository, q *ApiQuery) ([]byte, error) {
	tsyms, err := json.Marshal(q.Tsyms)
	if err != nil {
		fmt.Printf("Error: %m", err.Error())
	}
	fsyms, err := json.Marshal(q.Fsyms)
	if err != nil {
		fmt.Printf("Error: %m", err.Error())
	}

	result := r.FindByTsymsAndFsyms(string(fsyms), string(tsyms))

	if result == nil {
		return nil, fmt.Errorf("resource not available")
	}

	return []byte(result.Data), nil
}

func loadApiData(q *ApiQuery) ([]byte, error) {
	client := &http.Client{}
	fsyms := strings.Join(q.Fsyms[:], ",")
	tsyms := strings.Join(q.Tsyms[:], ",")
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("https://min-api.cryptocompare.com/data/pricemultifull?fsyms=%s&tsyms=%s", fsyms, tsyms),
		nil,
	)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//Convert the body to type string
	sb := string(body)
	var response ApiResponse
	err = json.Unmarshal([]byte(sb), &response)
	if err != nil {
		return nil, err
	}

	res, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func storeApiData(r *model.PriceRepository, q *ApiQuery, result []byte) {
	fsyms, _ := json.Marshal(q.Fsyms)
	tsyms, _ := json.Marshal(q.Tsyms)
	r.UpsertForTsymsAndFsyms(string(result), string(fsyms), string(tsyms))
}
