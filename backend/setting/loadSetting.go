package setting

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"path/filepath"
	"strconv"
)


var (
	// For global logger
	Logger *log.Logger

	clientConfig ClientSetting
	haveCreatedConfig bool = false
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
	LoggingLevel 		log.Level
	LoggingOutput 		string
	MaxActiveTorrents	int
	TorrentConfig		torrent.ClientConfig `json:"-"`
}

type ClientSetting struct {
	ConnectSetting
	EngineSetting
}

func (clientConfig *ClientSetting) loadDefaultConfig() {
	var ClientConfig ClientSetting
	ClientConfig.EngineSetting.LoggingLevel = log.WarnLevel
	ClientConfig.EngineSetting.TorrentConfig.DataDir = "download"
	ClientConfig.EngineSetting.TorrentConfig.Seed = false

	ClientConfig.ConnectSetting.SupportRemote = false
	ClientConfig.ConnectSetting.IP = "127.0.0.1"
	ClientConfig.ConnectSetting.Port = 8482
	ClientConfig.ConnectSetting.Addr = ClientConfig.ConnectSetting.IP + ":" + string(ClientConfig.ConnectSetting.Port)
	ClientConfig.ConnectSetting.AuthUsername = "ANT"
	ClientConfig.ConnectSetting.AuthPassword = "passwd"
}

// TODO
func calculateRateLimiters(uploadRate, downloadRate string) (*rate.Limiter, *rate.Limiter) {
	downloadRateLimiter := rate.NewLimiter(rate.Inf, 0)
	uploadRateLimiter := rate.NewLimiter(rate.Inf, 0)
	return uploadRateLimiter, downloadRateLimiter
}

func (clientConfig *ClientSetting) loadValueFromConfig() {

	clientConfig.EngineSetting.LoggingLevel  		= log.AllLevels[viper.GetInt("EngineSetting.LoggingLevel")]
	clientConfig.EngineSetting.LoggingOutput 		= viper.GetString("EngineSetting.LoggingOutput")
	clientConfig.EngineSetting.UseSocksproxy 		= viper.GetBool("EngineSetting.UseSocksproxy")
	clientConfig.EngineSetting.SocksProxyURL 		= viper.GetString("EngineSetting.SocksProxyURL")
	clientConfig.EngineSetting.MaxActiveTorrents 	= viper.GetInt("EngineSetting.MaxActiveTorrents")

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

	clientConfig.EngineSetting.TorrentConfig.UploadRateLimiter, clientConfig.EngineSetting.TorrentConfig.DownloadRateLimiter = calculateRateLimiters(viper.GetString("EngineSetting.TorrentConfig.UploadRateLimit"), viper.GetString("EngineSetting.TorrentConfig.DownloadRateLimit"))
	tmpDataDir, err := filepath.Abs(filepath.ToSlash(viper.GetString("EngineSetting.TorrentConfig.DataDir")))
	clientConfig.EngineSetting.TorrentConfig.DataDir = tmpDataDir
	if err != nil {
		fmt.Printf("Fail to create default datadir %s \n", err)
	}
	tmpListenAddr := viper.GetString("EngineSetting.TorrentConfig.ListenAddr")
	if tmpListenAddr != "" {
		clientConfig.EngineSetting.TorrentConfig.SetListenAddr(tmpListenAddr)
	}
	clientConfig.EngineSetting.TorrentConfig.ListenPort = clientConfig.ConnectSetting.Port
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

func (clientConfig *ClientSetting) UpdateConfig (updateKey, updateValue string)  {
	viper.Set(updateKey, updateValue)
	viper.WriteConfig()
}

// Load setting from config.toml
func createEngineSetting() ClientSetting{
	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("error config file : %S \n", err)
		fmt.Println("use default config")
		clientConfig.loadDefaultConfig()
	}else {
		clientConfig.loadValueFromConfig()
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		clientConfig.loadValueFromConfig()
	})
	return clientConfig
}

func GetClientSetting() ClientSetting {
	if haveCreatedConfig == false {
		haveCreatedConfig = true
		createEngineSetting()
	}
	return clientConfig
}












