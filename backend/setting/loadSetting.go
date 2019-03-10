package setting

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
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

}

// Load setting from config.toml
func (clientConfig *ClientSetting) createClientSetting()(){

	clientConfig.loadDefaultConfig()

	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
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












