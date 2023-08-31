package main

import (
	"crypto/md5"
	"fmt"
//	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/njhsi/atts.agi/internal/agi"
//	"github.com/njhsi/atts.agi/internal/oauth"
//	"github.com/njhsi/atts.agi/internal/speech"
)

type codec struct {
	codec string
	rate  string
}

const debug = false
const workDir = "/var/run/asterisk/"

var codecs = map[string]codec{
	"silk12":  codec{"sln12", "12000"},
	"sln12":   codec{"sln12", "12000"},
	"speex16": codec{"sln16", "16000"},
	"slin16":  codec{"sln16", "16000"},
	"silk16":  codec{"sln16", "16000"},
	"g722":    codec{"sln16", "16000"},
	"siren7":  codec{"sln16", "16000"},
	"speex32": codec{"sln32", "32000"},
	"slin32":  codec{"sln32", "32000"},
	"celt32":  codec{"sln32", "32000"},
	"siren14": codec{"sln32", "32000"},
	"celt44":  codec{"sln44", "44100"},
	"slin44":  codec{"sln44", "44100"},
	"celt48":  codec{"sln44", "44100"},
	"slin48":  codec{"sln44", "44100"},
	"other":   codec{"sln", "8000"},
}

//go build -ldflags "-X main.baiduSpeechAPIKey=xxx main.baiduSpeechSecretKey=xxx"
var baiduSpeechAPIKey string
var baiduSpeechSecretKey string

func main() {

	text := ""
	intkey := false

	log.Println(os.Args)

	if len(os.Args) == 2 {
		text = os.Args[1]
	} else if (len(os.Args) == 3) && (os.Args[2] == "true") {
		text = os.Args[1]
		intkey = true
	} else if (len(os.Args) == 3) && (os.Args[2] == "false") {

		text = os.Args[1]
		intkey = false
	} else {
		log.Fatalf("Your provided args are funky and incorrect, please check and correct")
	}

	// Create a new AGI session and Parse the AGI environment.
	myAgi := agi.New()
	err := myAgi.Init(nil)

	if err != nil {
		log.Fatalf("Error Parsing AGI environment: %v\n", err)
	}

	// Handle Hangup from the asterisk server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	go handleHangup(sigChan)

	//check/detect format
	myAgi.Verbose(myAgi.Env["channel"])
	foos, err := myAgi.GetFullVariable("${CHANNEL(audionativeformat)}", myAgi.Env["channel"])
	myAgi.Verbose(foos)
	myAgi.Verbose(foos.Dat)
	audioformat := "other"
	re := regexp.MustCompile("((silk|sln)12)|((speex|slin|silk)16|g722|siren7)|((speex|slin|celt)32|siren14)|((celt|slin)44)|((celt|slin)48)")
	if re.FindString(foos.Dat) != "" {
		audioformat = re.FindString(foos.Dat)
	}

	//Playback recieved text Message
	for i := 0; i < 10; i++ { //handleHangup goroutine might terminate this..
		myAgi.Verbose(playback(text, audioformat, intkey, myAgi))
	}
}

//play text to asterisk
func playback(text string, format string, intkey bool, myAgi *agi.Session) string {

	//store a hashed verstion of our message to use as a filename
	name := "TtS_"+hashString(text)
	//var to hold the name of the file we will be streaming
	astName := name

	//check if we already have this message cached in our temp directory and play that back
	//if already cached is false we'll generate a new sound file for playback
	if alreadyCached(name, format) == false {
		os.Remove(workDir + "TtS_*") 
		//mp3Name := getText2Speach( "pico message: "+text, name )
		//wavName := convert2Wav( mp3Name )
		wavName := getText2Speach_pico(text, name)
		astName = convert2Aster(wavName, format)
	} else {
		astName = name
	}

	//var to hold our errors
	var err error = nil
	//var to hold our replies
	var rep agi.Reply

	//if intkey is true we'll allow our message to be interrupted with a press of 1 or 2, else we'll play through
	if intkey == true {
		rep, err = myAgi.StreamFile(workDir+astName, "12")
		log.Println("you pressed: " + fmt.Sprintf("%c", rep.Res))
	} else {
		rep, err = myAgi.StreamFile(workDir+astName, "none")
		log.Println("pressed: " + fmt.Sprintf("%c", rep.Res))

	}

	//throw back a message if we encounter an error during playback
	if err != nil {
		log.Fatalf("AGI reply error: %v\n", err)
	}
	if rep.Res == -1 {
		log.Println("Error during playback\n")
	}

	//if we recieve a 1 or 2 to ack or escalate set the extension here
	if (rep.Res == 49) || (rep.Res == 50) {
		log.Println("returning extension since 1 or 2 was pressed: " + fmt.Sprintf("%c", rep.Res))
		myAgi.SetVariable("INTEXTEN", fmt.Sprintf("%c", rep.Res))
	}

	//return "empty" if we don't get an interuption/response
	return "empty"
}

//create a hash from the message and return a string we'll use for a filename
func hashString(text string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(text)))
}

//check if we already have an audio file cached with the string we're translating
func alreadyCached(filename string, audioformat string) bool {
	if _, err := os.Stat(workDir + filename + "." + codecs[audioformat].codec); os.IsNotExist(err) {
		return false
	}
	return true
}

func getText2Speach_pico(text string, filename string) string {                                 
        _, err := exec.Command("/usr/bin/pico2wave", "-w",  workDir+filename+".wav", text).Output()
                                                                  
        if err != nil {                    
                log.Fatal(err)             
        }     
        return filename                                                
}    

//do our text 2 speach dealie... retrieved as an mp3
/*func getText2Speach_baidu(text string, filename string) string {

	auth := oauth.NewBaiduOauth(baiduSpeechAPIKey, baiduSpeechSecretKey, oauth.NewCacheImp())
	tts := speech.NewBaiduTTS(auth)
	data, err := tts.ToSpeech(text)
	if err != nil {
		log.Fatal(err)
	}

	//write out our audio file, note as an mp3
	err = ioutil.WriteFile(workDir+filename+".wav", data, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	return filename
}
*/

//use mpg123 to convert our recieved mp3 to a wav
func convert2Wav(filename string) string {

	//mpg123 -q -w <wav_output_file_name> <file_to_convert>
	result, err := exec.Command("mpg123", "-q", "-w", workDir+filename+".wav", workDir+filename+".mp3").Output()

	if err != nil {
		log.Fatalf("error converting mp3 to wav: %v\n", err)
	} else if debug == true {
		log.Println(result)
	}
	return filename
}

//use sox to convert to an asterisk friendly sound format
func convert2Aster(filename string, audioformat string) string {

	result, err := exec.Command("/usr/bin/sox", workDir+filename+".wav", "-q", "-r", codecs[audioformat].rate, "-t", "raw", workDir+filename+"."+codecs[audioformat].codec).Output()

	if err != nil {
		log.Fatalf("error transcoding with sox: %v\n", err)
	} else {
		os.Remove(workDir + filename + ".mp3")
		os.Remove(workDir + filename + ".wav")
	}
	if debug == true {
		log.Println(result)
	}
	return filename
}

//hang up the line if we hang up
func handleHangup(sch <-chan os.Signal) {
	select {
	case signal := <-sch:
		log.Println("Received %v, exiting...\n", signal)
		os.Exit(1)
	case <-time.After(time.Second * 55):
		log.Println("Timeout 55 secs\n")
		os.Exit(0)
	}
}
