package conf

import (
	coreinformers "k8s.io/client-go/informers/core/v1"
	restclient "k8s.io/client-go/rest"
	clientSet "github.com/gxthrj/apisix-ingress-types/pkg/client/clientset/versioned"
	seven "github.com/gxthrj/seven/conf"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/informers"
	"os"
	"path/filepath"
	"io/ioutil"
	"fmt"
	"github.com/tidwall/gjson"
	"runtime"
)

var (
	ENV      string
	basePath string
	ADMIN_URL = os.Getenv("APISIX_ADMIN_INTERNAL")
	HOSTNAME = os.Getenv("HOSTNAME")
	LOCAL_ADMIN_URL = ""
	podInformer coreinformers.PodInformer
	svcInformer coreinformers.ServiceInformer
	nsInformer coreinformers.NamespaceInformer
	EndpointsInformer coreinformers.EndpointsInformer
	IsLeader = false
	//etcdClient client.Client
	kubeClient kubernetes.Interface
	CoreSharedInformerFactory informers.SharedInformerFactory
)
const PROD = "prod"
const HBPROD = "hb-prod"
const BETA = "beta"
const DEV = "dev"
const TEST = "test"
const LOCAL = "local"
const confPath = "/root/ingress-controller/conf.json"
const AispeechUpstreamKey = "/apisix/customer/upstream/map"

func setEnvironment() {
	if env := os.Getenv("ENV"); env == "" {
		ENV = LOCAL
	} else {
		ENV = env
	}
	_, basePath, _, _ = runtime.Caller(1)
}

func ConfPath() string {
	if ENV == LOCAL {
		return filepath.Join(filepath.Dir(basePath), "conf.json")
	} else {
		return confPath
	}
}

type etcdConfig struct {
	Addresses []string
}

var EtcdConfig etcdConfig
var K8sAuth k8sAuth
var Syslog syslog

var config *restclient.Config

func init() {
	// 获取当前环境
	setEnvironment()
	// 获取配置文件路径
	filePath := ConfPath()
	// 获取配置文件内容
	if configurationContent, err := ioutil.ReadFile(filePath); err != nil {
		panic(fmt.Sprintf("failed to read configuration file: %s", filePath))
	} else {
		configuration := gjson.ParseBytes(configurationContent)
		// apisix baseUrl
		apisixConf := configuration.Get("conf.apisix")
		apisixBaseUrl := apisixConf.Get("base_url").String()
		seven.SetBaseUrl(apisixBaseUrl)
		// k8sAuth conf
		k8sAuthConf := configuration.Get("conf.k8sAuth")
		K8sAuth.file = k8sAuthConf.Get("file").String()
		// syslog conf
		syslogConf := configuration.Get("conf.syslog")
		Syslog.Host = syslogConf.Get("host").String()
	}
	// init etcd client
	//etcdClient = NewEtcdClient()
	// init informer
	InitInformer()
}

type k8sAuth struct {
	file string
}

type syslog struct {
	Host string
}

//func GetEtcdAPI() client.KeysAPI{
//	return client.NewKeysAPI(etcdClient)
//}


func GetURL() string{
	if ADMIN_URL != "" {
		return ADMIN_URL
	} else {
		return "http://172.16.20.90:30116/apisix/admin"
	}
}

func GetPodInformer() coreinformers.PodInformer{
	return podInformer
}

func GetSvcInformer() coreinformers.ServiceInformer{
	return svcInformer
}

func GetNsInformer() coreinformers.NamespaceInformer{
	return nsInformer
}

func GetKubeClient() kubernetes.Interface{
	return kubeClient
}

func InitKubeClient() kubernetes.Interface {
	//var err error
	//if ENV == LOCAL {
	//	clientConfig, err := clientcmd.LoadFromFile(K8sAuth.file)
	//	ExceptNilErr(err)
	//
	//	config, err = clientcmd.NewDefaultClientConfig(*clientConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
	//	ExceptNilErr(err)
	//} else {
	//	config, err = restclient.InClusterConfig()
	//	ExceptNilErr(err)
	//}

	k8sClient, err := kubernetes.NewForConfig(config)
	ExceptNilErr(err)
	return k8sClient
}

func InitApisixClient() clientSet.Interface{
	apisixRouteClientset, err:= clientSet.NewForConfig(config)
	ExceptNilErr(err)
	return apisixRouteClientset
}

func InitInformer() {
	// 生成一个k8s client
	//var config *restclient.Config
	var err error
	if ENV == LOCAL {
		clientConfig, err := clientcmd.LoadFromFile(K8sAuth.file)
		ExceptNilErr(err)

		config, err = clientcmd.NewDefaultClientConfig(*clientConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
		ExceptNilErr(err)
	} else {
		config, err = restclient.InClusterConfig()
		ExceptNilErr(err)
	}

	//k8sClient, err := kubernetes.NewForConfig(config)
	kubeClient = InitKubeClient()
	ExceptNilErr(err)

	// 创建一个informerFactory
	//sharedInformerFactory := informers.NewSharedInformerFactory(k8sClient, 0)
	// 创建一个informerFactory
	CoreSharedInformerFactory = informers.NewSharedInformerFactory(kubeClient, 0)

	// 创建 informers
	podInformer = CoreSharedInformerFactory.Core().V1().Pods()
	svcInformer = CoreSharedInformerFactory.Core().V1().Services()
	nsInformer = CoreSharedInformerFactory.Core().V1().Namespaces()
	//return podInformer, svcInformer, nsInformer
}

func ExceptNilErr(err error)  {
	if err != nil {
		panic(err)
	}
}

//func NewEtcdClient() client.Client {
//	cfg := client.Config{
//		Endpoints: EtcdConfig.Addresses,
//		Transport: client.DefaultTransport,
//	}
//	if c, err := client.New(cfg); err != nil {
//		panic(fmt.Sprintf("failed to initialize etcd watcher. %s", err.Error()))
//	} else {
//		return c
//	}
//}

// EtcdWatcher
//type EtcdWatcher struct {
//	client     client.Client
//	etcdKey 	string
//	ctx        context.Context
//	cancels    []context.CancelFunc
//}
//
//
//type BalancerRules struct {
//	RuleSpec *RuleSpec `json:"spec"`
//}
//
//type RuleSpec struct {
//	Ewma []string `json:"ewma"`
//	Sllb []Sllb `json:"sllb"`
//}
//
//type Sllb struct {
//	Name string `json:"name"`
//	Threshold int64 `json:"threshold"`
//	Open string `json:"open"`
//	MakeZero string `json:"makeZero"`
//}
//
//type BalancerLevel struct {
//	LevelSpec *LevelSpec `json:"spec"`
//}
//
//type LevelSpec struct {
//	Pod []string `json:"pod"`
//}

