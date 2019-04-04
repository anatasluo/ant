package setting

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/iplist"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)


var (
	clientConfig ClientSetting
	haveCreatedConfig 				= false
	globalViper						= viper.New()
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
	UseSocksproxy 			bool
	SocksProxyURL 			string
	MaxActiveTorrents		int
	TorrentDBPath			string
	TorrentConfig			torrent.ClientConfig `json:"-"`
	Tmpdir					string
	MaxEstablishedConns 	int
	EnableDefaultTrackers 	bool
	DefaultTrackers			[][]string
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

// These settings can be determined by users
type WebSetting struct {
	UseSocksproxy 			bool
	SocksProxyURL 			string
	MaxEstablishedConns 	int
	Tmpdir					string
	DataDir					string
	EnableDefaultTrackers 	bool
	DefaultTrackerList		string
}

func (clientConfig *ClientSetting) GetWebSetting()(webSetting WebSetting)  {
	webSetting.EnableDefaultTrackers = clientConfig.EngineSetting.EnableDefaultTrackers
	webSetting.DefaultTrackerList = globalViper.GetString("EngineSetting.DefaultTrackerList")
	webSetting.UseSocksproxy = clientConfig.EngineSetting.UseSocksproxy
	webSetting.SocksProxyURL = clientConfig.EngineSetting.SocksProxyURL
	webSetting.MaxEstablishedConns = clientConfig.MaxEstablishedConns
	webSetting.Tmpdir = clientConfig.Tmpdir
	webSetting.DataDir = clientConfig.TorrentConfig.DataDir
	return
}

// TODO
func calculateRateLimiters(uploadRate, downloadRate string) (*rate.Limiter, *rate.Limiter) {
	downloadRateLimiter := rate.NewLimiter(rate.Inf, 0)
	uploadRateLimiter := rate.NewLimiter(rate.Inf, 0)
	return uploadRateLimiter, downloadRateLimiter
}

func (clientConfig *ClientSetting) loadValueFromConfig()() {

	clientConfig.LoggerSetting.LoggingLevel  		= log.AllLevels[globalViper.GetInt("LoggerSetting.LoggingLevel")]
	clientConfig.LoggerSetting.LoggingOutput 		= globalViper.GetString("LoggerSetting.LoggingOutput")
	clientConfig.LoggerSetting.Logger.SetLevel(clientConfig.LoggerSetting.LoggingLevel)

	clientConfig.EngineSetting.UseSocksproxy 		= globalViper.GetBool("EngineSetting.UseSocksproxy")
	clientConfig.EngineSetting.SocksProxyURL 		= globalViper.GetString("EngineSetting.SocksProxyURL")
	clientConfig.EngineSetting.MaxActiveTorrents 	= globalViper.GetInt("EngineSetting.MaxActiveTorrents")
	clientConfig.EngineSetting.TorrentDBPath 		= globalViper.GetString("EngineSetting.TorrentDBPath")
	clientConfig.EngineSetting.MaxEstablishedConns 	= globalViper.GetInt("EngineSetting.MaxEstablishedConns")
	tmpDir, tmpErr := filepath.Abs(filepath.ToSlash(globalViper.GetString("EngineSetting.Tmpdir")))
	_ = os.Mkdir(tmpDir, 0755)
	clientConfig.EngineSetting.Tmpdir				= tmpDir
	if tmpErr != nil {
		clientConfig.Logger.WithFields(log.Fields{"Error":tmpErr}).Error("Fail to create default cache directory")
	}

	clientConfig.ConnectSetting.IP = globalViper.GetString("ConnectSetting.IP")
	clientConfig.ConnectSetting.Port = globalViper.GetInt("ConnectSetting.Port")
	clientConfig.ConnectSetting.SupportRemote = globalViper.GetBool("ConnectSetting.SupportRemote")
	if clientConfig.ConnectSetting.SupportRemote {
		clientConfig.ConnectSetting.Addr = ":" + strconv.Itoa(clientConfig.ConnectSetting.Port)
	} else {
		clientConfig.ConnectSetting.Addr = clientConfig.ConnectSetting.IP + ":" + strconv.Itoa(clientConfig.ConnectSetting.Port)
	}
	clientConfig.ConnectSetting.AuthUsername = globalViper.GetString("ConnectSetting.AuthUsername")
	clientConfig.ConnectSetting.AuthPassword = globalViper.GetString("ConnectSetting.AuthPassword")

	clientConfig.EngineSetting.TorrentConfig = *torrent.NewDefaultClientConfig()
	clientConfig.EngineSetting.TorrentConfig.UploadRateLimiter, clientConfig.EngineSetting.TorrentConfig.DownloadRateLimiter = calculateRateLimiters(viper.GetString("TorrentConfig.UploadRateLimit"), viper.GetString("TorrentConfig.DownloadRateLimit"))
	tmpDataDir, err := filepath.Abs(filepath.ToSlash(globalViper.GetString("EngineSetting.DataDir")))
	_ = os.Mkdir(tmpDataDir, 0755)
	clientConfig.EngineSetting.TorrentConfig.DataDir = tmpDataDir
	if err != nil {
		clientConfig.Logger.WithFields(log.Fields{"Error":err}).Error("Fail to create default datadir")
	}
	tmpListenAddr := globalViper.GetString("TorrentConfig.ListenAddr")
	if tmpListenAddr != "" {
		clientConfig.EngineSetting.TorrentConfig.SetListenAddr(tmpListenAddr)
	}
	clientConfig.EngineSetting.TorrentConfig.ListenPort = clientConfig.ConnectSetting.Port
	clientConfig.EngineSetting.TorrentConfig.ListenPort = globalViper.GetInt("TorrentConfig.ListenPort")
	clientConfig.EngineSetting.TorrentConfig.DisablePEX = globalViper.GetBool("TorrentConfig.DisablePEX")
	clientConfig.EngineSetting.TorrentConfig.NoDHT = globalViper.GetBool("TorrentConfig.NoDHT")
	clientConfig.EngineSetting.TorrentConfig.NoUpload = globalViper.GetBool("TorrentConfig.NoUpload")
	clientConfig.EngineSetting.TorrentConfig.Seed = globalViper.GetBool("TorrentConfig.Seed")
	clientConfig.EngineSetting.TorrentConfig.DisableUTP = globalViper.GetBool("TorrentConfig.DisableUTP")
	clientConfig.EngineSetting.TorrentConfig.DisableTCP = globalViper.GetBool("TorrentConfig.DisableTCP")
	clientConfig.EngineSetting.TorrentConfig.DisableIPv6 = globalViper.GetBool("TorrentConfig.DisableIPv6")
	clientConfig.EngineSetting.TorrentConfig.Debug = globalViper.GetBool("TorrentConfig.Debug")
	clientConfig.EngineSetting.TorrentConfig.PeerID = globalViper.GetString("TorrentConfig.PeerID")

	clientConfig.EngineSetting.TorrentConfig.EncryptionPolicy.DisableEncryption = globalViper.GetBool("EncryptionPolicy.DisableEncryption")
	clientConfig.EngineSetting.TorrentConfig.EncryptionPolicy.ForceEncryption = globalViper.GetBool("EncryptionPolicy.ForceEncryption")
	clientConfig.EngineSetting.TorrentConfig.EncryptionPolicy.PreferNoEncryption = globalViper.GetBool("EncryptionPolicy.PreferNoEncryption")

	//blocklistPath, err := filepath.Abs("biglist.p2p.gz")
	//if err != nil {
	//	fmt.Printf("Failed to update block list, Err is %s\n", err)
	//}
	//clientConfig.EngineSetting.TorrentConfig.IPBlocklist = getBlocklist(blocklistPath, globalViper.GetString("EngineSetting.DefaultIPBlockList"))

	clientConfig.EngineSetting.EnableDefaultTrackers = globalViper.GetBool("EngineSetting.EnableDefaultTrackers")
	if clientConfig.EngineSetting.EnableDefaultTrackers {
		trackerPath, err := filepath.Abs("tracker.txt")
		if err != nil {
			clientConfig.Logger.WithFields(log.Fields{"Error":err}).Error("Failed to update trackers list")
		}
		clientConfig.EngineSetting.DefaultTrackers = clientConfig.getDefaultTrackers(trackerPath, globalViper.GetString("EngineSetting.DefaultTrackerList"))
	} else {
		clientConfig.EngineSetting.DefaultTrackers = [][]string{}
	}

	if clientConfig.UseSocksproxy {
		clientConfig.TorrentConfig.ProxyURL = clientConfig.SocksProxyURL
	}

	if clientConfig.LoggerSetting.LoggingOutput == "file" {
		file, err := os.OpenFile(clientConfig.TorrentConfig.DataDir + "/ant_engine.log", os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			clientConfig.Logger.WithFields(log.Fields{"Error":err}).Error("Failed to open log file")
		} else {
			clientConfig.Logger.Out = file
		}
	}else{
		clientConfig.Logger.Out = os.Stdout
	}
}

func (clientConfig *ClientSetting) loadFromConfigFile()() {
	globalViper.SetConfigName("config")
	globalViper.AddConfigPath("./")
	err := globalViper.ReadInConfig()
	if err != nil {
		clientConfig.LoggerSetting.Logger.WithFields(log.Fields{"Detail":err}).Fatal("Can not find config.toml")
	}else {
		clientConfig.loadValueFromConfig()
	}
	globalViper.WatchConfig()
	//globalViper.OnConfigChange(func(e fsnotify.Event) {
	//	fmt.Println("Config file changed:", e.Name)
	//	//clientConfig.loadValueFromConfig()
	//})
}

// Load setting from config.toml
func (clientConfig *ClientSetting) createClientSetting()(){

	//Default settings
	clientConfig.LoggerSetting.Logger = log.New()

	clientConfig.loadFromConfigFile()
}

func GetClientSetting() *ClientSetting {
	if haveCreatedConfig == false {
		haveCreatedConfig = true
		clientConfig.createClientSetting()
	}
	return &clientConfig
}


func (clientConfig *ClientSetting) UpdateConfig (newSetting WebSetting)()  {
	globalViper.Set("EngineSetting.EnableDefaultTrackers", newSetting.EnableDefaultTrackers)
	globalViper.Set("EngineSetting.DefaultTrackerList", newSetting.DefaultTrackerList)
	globalViper.Set("EngineSetting.UseSocksproxy", newSetting.UseSocksproxy)
	globalViper.Set("EngineSetting.SocksProxyURL", newSetting.SocksProxyURL)
	globalViper.Set("EngineSetting.MaxEstablishedConns", newSetting.MaxEstablishedConns)
	globalViper.Set("EngineSetting.Tmpdir", newSetting.Tmpdir)
	globalViper.Set("EngineSetting.DataDir", newSetting.DataDir)

	tr, err := toml.TreeFromMap(globalViper.AllSettings())
	trS := tr.String()
	err = ioutil.WriteFile("config.toml", []byte(trS), 0644)
	if err != nil {
		clientConfig.Logger.WithFields(log.Fields{"Error": err, "Settings": trS}).Fatal("Unable to update settings")
	}
	haveCreatedConfig = false
	GetClientSetting()
}


func (clientConfig *ClientSetting) getDefaultTrackers(filepath string, url string) [][]string {
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
	clientConfig.downloadFile(url, filepath)
	return res
}

// Download and add the blocklist.
func (clientConfig *ClientSetting) getBlocklist(filepath string, blocklistURL string) iplist.Ranger {

	// Load blocklist.
	// #nosec
	// We trust our temporary directory as we just wrote the file there ourselves.
	blocklistReader, err := os.Open(filepath)
	if err != nil {
		clientConfig.Logger.WithFields(log.Fields{"Error":err}).Error("Error opening blocklist")
		return nil
	}

	// Extract file.
	gzipReader, err := gzip.NewReader(blocklistReader)
	if err != nil {
		clientConfig.Logger.WithFields(log.Fields{"Error":err}).Error("Error extracting blocklist")
		return nil
	}

	// Read as iplist.
	blocklist, err := iplist.NewFromReader(gzipReader)
	if err != nil {
		clientConfig.Logger.WithFields(log.Fields{"Error":err}).Error("Error reading blocklist")
		return nil
	}
	clientConfig.Logger.Debug("Loading blocklist")

	//Update list if possible for next time
	clientConfig.downloadFile(blocklistURL, filepath)
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
	defer func() {
		_ = file.Close()
	}()

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

func (clientConfig *ClientSetting) downloadFile(downloadURL string, filepath string) {

	go func() {
		// Get the data
		resp, err := http.Get(downloadURL)
		if err != nil {
			clientConfig.Logger.WithFields(log.Fields{"Error":err}).Error("Failed to update list")
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		// Create the file
		out, err := os.Create(filepath)
		defer func() {
			_ = out.Close()
		}()
		if err != nil {
			if err != nil {
				clientConfig.Logger.WithFields(log.Fields{"Error":err}).Error("Failed to create list")
				return
			}
		}


		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			if err != nil {
				clientConfig.Logger.WithFields(log.Fields{"Error":err}).Error("Failed to trackers list")
				return
			}
		}
		clientConfig.Logger.Info("update tracker list successfully")
	}()
}



