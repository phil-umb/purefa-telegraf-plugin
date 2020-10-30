package purefa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// PureFa - plugin main structure
type PureFA struct {
	Array    string `toml:"array"`
	APIToken string `toml:"api_token"`
	//APIVersion
	HTTPTimeout internal.Duration `toml:"http_timeout"`
	client      *http.Client
}

const sampleConfig = `
  ## Arrays to monitor, with token for each array.
  array = "purefa1.example.com"
  ## Pure API consumer key.  One for each array
  # api_token = "55b0bf09-10b4-e54c-7db8-9fccc742e908"
  ## HTTP Timeout
  # http_timeout = 10
  ## Filter Snapshots - filter out Veeam created volume copies
  # ignore_veeamsnap = true
  #Multiple arrays with different tokens can be specified. @TODO
`

// SampleConfig returns sample configuration for this plugin.
func (t *PureFA) SampleConfig() string {
	return sampleConfig
}

// Description returns the plugin description.
func (t *PureFA) Description() string {
	return "Gather performance and capacity information from Pure Storage FlashArrays."
}

// Get a new Rest CLient and all the infos
func NewPureFA() *PureFA {
	tr := &http.Transport{ResponseHeaderTimeout: time.Duration(3 * time.Second)}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(4 * time.Second),
	}
	return &PureFA{client: client}
}

type pureVols struct {
	Created string `json:"created"`
	Name    string `json:"name"`
	Serial  string `json:"serial"`
	Size    int64  `json:"size"`
}

// Init the things
func (t *PureFA) Init() error {

	if len(t.APIToken) == 0 {
		return fmt.Errorf("You must specify an API Token")
	}
	if len(t.URL) == 0 {
		t.URL = "https://" + t.Array + "/api/1.15"
	}
	// Have a default timeout of 4s
	if t.HTTPTimeout.Duration == 0 {
		t.HTTPTimeout.Duration = time.Second * 4
	}

	t.client.Timeout = t.HTTPTimeout.Duration

	return nil
}

// Gather Pure FlashArray Metrics
func (t *PureFA) Gather(acc telegraf.Accumulator) error {
	// Perform GET request to get volumes
	req, err := http.NewRequest("GET", t.URL+"/volume", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+r.AuthToken)
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Successful responses will always return status code 200
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden {
			return fmt.Errorf("Pure Array responded with %d [Forbidden], verify your authToken", resp.StatusCode)
		}
		return fmt.Errorf("Pure Array responded with unexpected status code %d", resp.StatusCode)
	}
	// Decode the response JSON into a new stats struct
	var vols []pureVols
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return fmt.Errorf("unable to decode Pure volume response: %s", err)
	}
	// Range over all volumes, gathering stats. Returns early in case of any error.
	for _, v := range vols {
		t.gatherCapacity(v, acc)
		t.gatherPerformance(v, acc)
	}

	return nil
}

func (t *PureFA) gatherCapacity(v pureVols, acc telegraf.Accumulator) {
	tags := map[string]string{
		"name": v.Name,
	}
	fields
	acc.AddFields("purefa", fields, tags)
}

func init() {
	inputs.Add("purefa", func() telegraf.Input {
		return NewPureFA()
	})
}
