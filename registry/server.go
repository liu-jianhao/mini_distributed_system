package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const ServerPort = ":8020"
const ServicesURL = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []*Registration
	mutex         *sync.RWMutex
}

func (r *registry) add(reg *Registration) error {
	r.mutex.Lock()
	r.registrations = append(r.registrations, reg)
	r.mutex.Unlock()

	err := r.sendRequiredServices(reg)
	if err != nil {
		return err
	}

	r.notify(&patch{
		Added: []*patchEntry{
			{
				Name: reg.ServiceName,
				URL:  reg.ServiceURL,
			},
		},
	})

	return nil
}

func (r *registry) notify(pat *patch) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, reg := range r.registrations {
		go func(reg *Registration) {
			for _, reqService := range reg.RequiredServices {
				p := patch{
					Added:   []*patchEntry{},
					Removed: []*patchEntry{},
				}
				sendUpdate := false
				for _, added := range pat.Added {
					if added.Name == reqService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}
				for _, removed := range pat.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				if sendUpdate {
					err := r.sendPatch(p, reg.ServiceUpdateURL)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}
}

func (r *registry) remove(url string) error {
	for i, registration := range reg.registrations {
		if registration.ServiceURL == url {
			r.notify(&patch{
				Removed: []*patchEntry{
					{
						Name: registration.ServiceName,
						URL:  registration.ServiceURL,
					},
				},
			})
			r.mutex.Lock()
			reg.registrations = append(reg.registrations[:i], reg.registrations[i+1:]...)
			r.mutex.Unlock()
			return nil
		}
	}
	return fmt.Errorf("service at URL %s not found", url)
}

func (r *registry) heartbeat(t time.Duration) {
	for {
		var wg sync.WaitGroup
		for _, reg := range r.registrations {
			wg.Add(1)
			go func(reg *Registration) {
				defer wg.Done()
				success := true
				for attemps := 0; attemps < 3; attemps++ {
					res, err := http.Get(reg.HeartbeatURL)
					if err != nil {
						log.Println(err)
					} else if res.StatusCode == http.StatusOK {
						log.Printf("Heartbeat check passed for %v", reg.ServiceName)
						if !success {
							_ = r.add(reg)
						}
						break
					}
					log.Printf("Heartbeat check failed for %v", reg.ServiceName)
					if success {
						success = false
						_ = r.remove(reg.ServiceURL)
					}
					time.Sleep(1 * time.Second)
				}
			}(reg)
			wg.Wait()
			time.Sleep(t)
		}
	}
}

var once sync.Once

func SetupHeartbeat() {
	once.Do(func() {
		go reg.heartbeat(3 * time.Second)
	})
}

func (r *registry) sendRequiredServices(reg *Registration) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	p := patch{}
	for _, serviceReg := range r.registrations {
		for _, reqService := range reg.RequiredServices {
			if serviceReg.ServiceName != reqService {
				continue
			}
			p.Added = append(p.Added, &patchEntry{
				Name: serviceReg.ServiceName,
				URL:  serviceReg.ServiceURL,
			})
		}
	}

	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}

	return nil
}

func (r *registry) sendPatch(p patch, url string) error {
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}

	return nil
}

var reg = registry{
	registrations: make([]*Registration, 0),
	mutex:         new(sync.RWMutex),
}

type RegistryService struct{}

func (rs *RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var registration Registration
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&registration)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Adding Service name=%v, url=%s", registration.ServiceName, registration.ServiceURL)

		err = reg.add(&registration)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		url := string(payload)
		log.Printf("Removing service at URL: %s", url)
		err = reg.remove(url)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
