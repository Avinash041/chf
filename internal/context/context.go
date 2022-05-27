package context

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/free5gc/CDRUtil/cdrType"
	"github.com/free5gc/chf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

var chfCtx *CHFContext

func init() {
	chfCtx = new(CHFContext)
	chfCtx.Name = "chf"
	chfCtx.UriScheme = models.UriScheme_HTTPS
	chfCtx.NfService = make(map[models.ServiceName]models.NfService)
}

type CHFContext struct {
	NfId                      string
	Name                      string
	UriScheme                 models.UriScheme
	BindingIPv4               string
	RegisterIPv4              string
	SBIPort                   int
	NfService                 map[models.ServiceName]models.NfService
	NrfUri                    string
	LocalRecordSequenceNumber uint64
	UePool                    sync.Map
	ChargingSession           map[string]*cdrType.CHFRecord
}

// Create new CHF context
func CHF_Self() *CHFContext {
	return chfCtx
}

func (c *CHFContext) GetIPv4Uri() string {
	return fmt.Sprintf("%s://%s:%d", c.UriScheme, c.RegisterIPv4, c.SBIPort)
}

// Init NfService with supported service list ,and version of services
func (c *CHFContext) InitNFService(serviceList []factory.Service, version string) {
	tmpVersion := strings.Split(version, ".")
	versionUri := "v" + tmpVersion[0]
	for index, service := range serviceList {
		name := models.ServiceName(service.ServiceName)
		c.NfService[name] = models.NfService{
			ServiceInstanceId: strconv.Itoa(index),
			ServiceName:       name,
			Versions: &[]models.NfServiceVersion{
				{
					ApiFullVersion:  version,
					ApiVersionInUri: versionUri,
				},
			},
			Scheme:          c.UriScheme,
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			ApiPrefix:       c.GetIPv4Uri(),
			IpEndPoints: &[]models.IpEndPoint{
				{
					Ipv4Address: c.RegisterIPv4,
					Transport:   models.TransportProtocol_TCP,
					Port:        int32(c.SBIPort),
				},
			},
			SupportedFeatures: service.SuppFeat,
		}
	}
}

// Allocate CHF Ue with supi and add to chf Context and returns allocated ue
func (c *CHFContext) NewCHFUe(Supi string) (*ChfUe, error) {
	if _, ok := c.ChfUeFindBySupi(Supi); ok {
		return nil, fmt.Errorf("Ue exist")
	}
	if strings.HasPrefix(Supi, "imsi-") {
		newUeContext := &ChfUe{}
		newUeContext.Supi = Supi
		c.UePool.Store(Supi, newUeContext)
		return newUeContext, nil
	} else {
		return nil, fmt.Errorf(" add Ue context fail ")
	}
}

func (context *CHFContext) ChfUeFindBySupi(supi string) (*ChfUe, bool) {
	if value, ok := context.UePool.Load(supi); ok {
		return value.(*ChfUe), ok
	}
	return nil, false
}
