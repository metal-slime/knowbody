package knowbody

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
)

var (
	CurrentConfig Config
	State         CurrentState
)

// Start starts the loop to look at all the things
func Start() {
	CurrentConfig.SlackToken = os.Getenv("SLACK_TOKEN")

	State.Streams = make(map[string]ContentState)
	State.Channels = make(map[string]string)

	// Recover state from previous run
	ReadState()

	// Assume if LastRun is over a year old to not bother and just set it to now
	// Prevents spamming
	if State.LastRun.Before(time.Now().AddDate(-1, 0, 0)) {
		State.LastRun = time.Now()
	}

	for {
		// Allow config changes between runs
		ReadConfig()

		State.slackClient = slack.New(CurrentConfig.SlackToken)

		channels, err := State.slackClient.GetChannels(true)
		if err != nil {
			log.Fatalf("error getting slack channels: %s", err)
		}

		for _, channel := range channels {
			State.Channels[channel.Name] = channel.ID
		}

		for _, contentStream := range CurrentConfig.Streams {
			contentStream.Process()
		}

		State.LastRun = time.Now()

		WriteState()

		log.Print("Time to sleep for 60")

		time.Sleep(60 * time.Second)
	}
}

func Lint() {
	readYamlIntoConfig("conf.yaml", &CurrentConfig)
	readYamlIntoConfig("knowbody.lock", &State)
}

func ReadConfig() {
	err := DownloadFile("conf.yaml", "https://raw.githubusercontent.com/metal-slime/knowbody/master/conf.yaml")
	if err != nil {
		log.Printf("Error downloading updated config from Github: %s", err.Error())
	}

	readYamlIntoConfig("conf.yaml", &CurrentConfig)

	for key, stream := range CurrentConfig.Streams {
		comp, err := regexp.Compile(stream.Include)
		if err != nil {
			log.Fatalf("Error compiling regex '%s': %s", stream.Include, err.Error())
		}
		CurrentConfig.Streams[key].includeRegex = comp

		comp, err = regexp.Compile(stream.Exclude)
		if err != nil {
			log.Fatalf("Error compiling regex `%s`: %s", stream.Exclude, err.Error())
		}
		CurrentConfig.Streams[key].excludeRegex = comp
	}
}

func WriteState() {
	d, err := yaml.Marshal(&State)
	if err != nil {
		log.Fatalf("Error marshalling YAML: %s", err.Error())
	}

	err = ioutil.WriteFile("knowbody.lock", d, 0644)
	if err != nil {
		log.Fatalf("Error writing state file: %s", err.Error())
	}
}

func ReadState() {
	yamlFile, err := ioutil.ReadFile("knowbody.lock")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &State)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Error downloading file: Status Code %d", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func readYamlIntoConfig(file string, obj interface{}) {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, obj)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
}
