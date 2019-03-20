package setting

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/iplist"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)


var (
	clientConfig ClientSetting
	haveCreatedConfig 				= false
)

type ConnectSetting struct {
	SupportRemote 		bool
	IP					string
	Port 				int
	Addr				string
	AuthUsername 		string
	AuthPassword 		string
}

type EngineSetting struct {
	UseSocksproxy 		bool
	SocksProxyURL 		string
	MaxActiveTorrents	int
	TorrentDBPath		string
	TorrentConfig		torrent.ClientConfig `json:"-"`
	Tmpdir				string
	MaxEstablishedConns int
	DefaultTrackers		[][]string
}

type LoggerSetting struct {
	LoggingLevel 		log.Level
	LoggingOutput 		string
	Logger				*log.Logger
}

type ClientSetting struct {
	ConnectSetting
	EngineSetting
	LoggerSetting
}

func (clientConfig *ClientSetting) loadDefaultConfig()() {

	//TODO toml file to map[string]string and to json and then unmarshall to struct
	clientConfig.LoggerSetting.Logger = log.New()


}

// TODO
func calculateRateLimiters(uploadRate, downloadRate string) (*rate.Limiter, *rate.Limiter) {
	downloadRateLimiter := rate.NewLimiter(rate.Inf, 0)
	uploadRateLimiter := rate.NewLimiter(rate.Inf, 0)
	return uploadRateLimiter, downloadRateLimiter
}

func (clientConfig *ClientSetting) loadValueFromConfig()() {

	clientConfig.LoggerSetting.LoggingLevel  		= log.AllLevels[viper.GetInt("LoggerSetting.LoggingLevel")]
	clientConfig.LoggerSetting.LoggingOutput 		= viper.GetString("LoggerSetting.LoggingOutput")
	clientConfig.LoggerSetting.Logger.SetLevel(clientConfig.LoggerSetting.LoggingLevel)

	clientConfig.EngineSetting.UseSocksproxy 		= viper.GetBool("EngineSetting.UseSocksproxy")
	clientConfig.EngineSetting.SocksProxyURL 		= viper.GetString("EngineSetting.SocksProxyURL")
	clientConfig.EngineSetting.MaxActiveTorrents 	= viper.GetInt("EngineSetting.MaxActiveTorrents")
	clientConfig.EngineSetting.TorrentDBPath 		= viper.GetString("EngineSetting.TorrentDBPath")
	clientConfig.EngineSetting.MaxEstablishedConns 	= viper.GetInt("EngineSetting.MaxEstablishedConns")
	tmpDir, tmpErr := filepath.Abs(filepath.ToSlash(viper.GetString("EngineSetting.Tmpdir")))
	os.Mkdir(tmpDir, 0755)
	clientConfig.EngineSetting.Tmpdir				= tmpDir
	if tmpErr != nil {
		fmt.Printf("Fail to create default tmpdir %s \n", tmpErr)
	}

	clientConfig.ConnectSetting.IP = viper.GetString("ConnectSetting.IP")
	clientConfig.ConnectSetting.Port = viper.GetInt("ConnectSetting.Port")
	clientConfig.ConnectSetting.SupportRemote = viper.GetBool("ConnectSetting.SupportRemote")
	if clientConfig.ConnectSetting.SupportRemote {
		clientConfig.ConnectSetting.Addr = ":" + strconv.Itoa(clientConfig.ConnectSetting.Port)
	} else {
		clientConfig.ConnectSetting.Addr = clientConfig.ConnectSetting.IP + ":" + strconv.Itoa(clientConfig.ConnectSetting.Port)
	}
	clientConfig.ConnectSetting.AuthUsername = viper.GetString("ConnectSetting.AuthUsername")
	clientConfig.ConnectSetting.AuthPassword = viper.GetString("ConnectSetting.AuthPassword")

	clientConfig.EngineSetting.TorrentConfig = *torrent.NewDefaultClientConfig()
	clientConfig.EngineSetting.TorrentConfig.UploadRateLimiter, clientConfig.EngineSetting.TorrentConfig.DownloadRateLimiter = calculateRateLimiters(viper.GetString("EngineSetting.TorrentConfig.UploadRateLimit"), viper.GetString("EngineSetting.TorrentConfig.DownloadRateLimit"))
	tmpDataDir, err := filepath.Abs(filepath.ToSlash(viper.GetString("EngineSetting.TorrentConfig.DataDir")))
	os.Mkdir(tmpDataDir, 0755)
	clientConfig.EngineSetting.TorrentConfig.DataDir = tmpDataDir
	if err != nil {
		fmt.Printf("Fail to create default datadir %s \n", err)
	}
	tmpListenAddr := viper.GetString("EngineSetting.TorrentConfig.ListenAddr")
	if tmpListenAddr != "" {
		clientConfig.EngineSetting.TorrentConfig.SetListenAddr(tmpListenAddr)
	}
	//clientConfig.EngineSetting.TorrentConfig.ListenPort = clientConfig.ConnectSetting.Port
	clientConfig.EngineSetting.TorrentConfig.ListenPort = viper.GetInt("EngineSetting.TorrentConfig.ListenPort")
	clientConfig.EngineSetting.TorrentConfig.DisablePEX = viper.GetBool("EngineSetting.TorrentConfig.DisablePEX")
	clientConfig.EngineSetting.TorrentConfig.NoDHT = viper.GetBool("EngineSetting.TorrentConfig.NoDHT")
	clientConfig.EngineSetting.TorrentConfig.NoUpload = viper.GetBool("EngineSetting.TorrentConfig.NoUpload")
	clientConfig.EngineSetting.TorrentConfig.Seed = viper.GetBool("EngineSetting.TorrentConfig.Seed")
	clientConfig.EngineSetting.TorrentConfig.DisableUTP = viper.GetBool("EngineSetting.TorrentConfig.DisableUTP")
	clientConfig.EngineSetting.TorrentConfig.DisableTCP = viper.GetBool("EngineSetting.TorrentConfig.DisableTCP")
	clientConfig.EngineSetting.TorrentConfig.DisableIPv6 = viper.GetBool("EngineSetting.TorrentConfig.DisableIPv6")
	clientConfig.EngineSetting.TorrentConfig.Debug = viper.GetBool("EngineSetting.TorrentConfig.Debug")
	clientConfig.EngineSetting.TorrentConfig.PeerID = viper.GetString("EngineSetting.TorrentConfig.PeerID")

	clientConfig.EngineSetting.TorrentConfig.EncryptionPolicy.DisableEncryption = viper.GetBool("EngineSetting.TorrentConfig.EncryptionPolicy.DisableEncryption")
	clientConfig.EngineSetting.TorrentConfig.EncryptionPolicy.ForceEncryption = viper.GetBool("EngineSetting.TorrentConfig.EncryptionPolicy.ForceEncryption")
	clientConfig.EngineSetting.TorrentConfig.EncryptionPolicy.PreferNoEncryption = viper.GetBool("EngineSetting.TorrentConfig.EncryptionPolicy.PreferNoEncryption")

	//blocklistPath, err := filepath.Abs("biglist.p2p.gz")
	//if err != nil {
	//	fmt.Printf("Failed to update block list, Err is %s\n", err)
	//}
	//clientConfig.EngineSetting.TorrentConfig.IPBlocklist = getBlocklist(blocklistPath, viper.GetString("EngineSetting.TorrentConfig.defaultIPBlockList"))

	trackerPath, err := filepath.Abs("tracker.txt")
	if err != nil {
		fmt.Printf("Failed to update trackers list, Err is %s\n", err)
	}
	clientConfig.DefaultTrackers = getDefaultTrackers(trackerPath, viper.GetString("EngineSetting.TorrentConfig.defaultTrackerList"))
}

// Load setting from config.toml
func (clientConfig *ClientSetting) createClientSetting()(){

	clientConfig.loadDefaultConfig()

	viper.SetConfigName("config")

	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		clientConfig.LoggerSetting.Logger.WithFields(log.Fields{"Detail":err}).Fatal("Can not find config.toml")
	}else {
		clientConfig.loadValueFromConfig()
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		clientConfig.loadValueFromConfig()
	})

	//More Config for client
}

func GetClientSetting() ClientSetting {
	if haveCreatedConfig == false {
		haveCreatedConfig = true
		clientConfig.createClientSetting()
	}
	return clientConfig
}

//TODO : Not only update the config.toml, but also the clientconfig
func UpdateConfig (updateKey, updateValue string)()  {
	viper.Set(updateKey, updateValue)
	viper.WriteConfig()
}


func getDefaultTrackers(filepath string, url string) [][]string {
	datas, err := readLines(filepath)
	if err != nil {
		panic(err)
	}
	var res [][]string

	for i := range datas {
		if datas[i] != "" {
			res = append(res, []string{
				datas[i],
			})
		}
	}

	//Update list if possible for next time
	downloadFile(url, filepath)
	return res
}

// Download and add the blocklist.
func getBlocklist(filepath string, blocklistURL string) iplist.Ranger {

	// Load blocklist.
	// #nosec
	// We trust our temporary directory as we just wrote the file there ourselves.
	blocklistReader, err := os.Open(filepath)
	if err != nil {
		log.Printf("Error opening blocklist: %s", err)
		return nil
	}

	// Extract file.
	gzipReader, err := gzip.NewReader(blocklistReader)
	if err != nil {
		log.Printf("Error extracting blocklist: %s", err)
		return nil
	}

	// Read as iplist.
	blocklist, err := iplist.NewFromReader(gzipReader)
	if err != nil {
		log.Printf("Error reading blocklist: %s", err)
		return nil
	}

	log.Printf("Loading blocklist.\nFound %d ranges\n", blocklist.NumRanges())

	//Update list if possible for next time
	downloadFile(blocklistURL, filepath)
	return blocklist
}

// Read a whole file into the memory and store it as array of lines
func readLines(path string) (lines []string, err error) {
	var (
		file *os.File
		part []byte
		prefix bool
	)
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			lines = append(lines, buffer.String())
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func downloadFile(downloadURL string, filepath string) {

	go func() {
		// Get the data
		resp, err := http.Get(downloadURL)
		if err != nil {
			fmt.Printf("Failed to update list, Err is %s\n", err)
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			if err != nil {
				fmt.Printf("Failed to create list, Err is %s\n", err)
			}
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			if err != nil {
				fmt.Printf("Failed to trackers list, Err is %s\n", err)
			}
		}
	}()
}



