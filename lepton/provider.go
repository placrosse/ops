package lepton

import (
	"fmt"
	"os"
	"strings"

	"github.com/nanovms/ops/config"
)

var (
	// ErrInstanceNotFound is used when an instance doesn't exist in provider
	ErrInstanceNotFound = func(id string) error { return fmt.Errorf("Instance with id %v not found", id) }
)

var (
	// TTLDefault is the default ttl value used to create DNS records
	TTLDefault = 300
)

// Provider is an interface that provider must implement
type Provider interface {
	Initialize(config *config.ProviderConfig) error

	BuildImage(ctx *Context) (string, error)
	BuildImageWithPackage(ctx *Context, pkgpath string) (string, error)
	CreateImage(ctx *Context, imagePath string) error
	ListImages(ctx *Context) error
	GetImages(ctx *Context) ([]CloudImage, error)
	DeleteImage(ctx *Context, imagename string) error
	ResizeImage(ctx *Context, imagename string, hbytes string) error
	SyncImage(config *config.Config, target Provider, imagename string) error
	CustomizeImage(ctx *Context) (string, error)

	CreateInstance(ctx *Context) error
	ListInstances(ctx *Context) error
	GetInstances(ctx *Context) ([]CloudInstance, error)
	GetInstanceByID(ctx *Context, id string) (*CloudInstance, error)
	DeleteInstance(ctx *Context, instancename string) error
	StopInstance(ctx *Context, instancename string) error
	StartInstance(ctx *Context, instancename string) error
	GetInstanceLogs(ctx *Context, instancename string) (string, error)
	PrintInstanceLogs(ctx *Context, instancename string, watch bool) error

	VolumeService
}

// Storage is an interface that provider's storage must implement
type Storage interface {
	CopyToBucket(config *config.Config, source string) error
}

// VolumeService is an interface for volume related operations
type VolumeService interface {
	CreateVolume(ctx *Context, name, data, size, provider string) (NanosVolume, error)
	GetAllVolumes(ctx *Context) (*[]NanosVolume, error)
	DeleteVolume(ctx *Context, name string) error
	AttachVolume(ctx *Context, image, name, mount string) error
	DetachVolume(ctx *Context, image, name string) error
}

// DNSRecord is ops representation of a dns record
type DNSRecord struct {
	Name string
	IP   string
	Type string
	TTL  int
}

// DNSService is an interface for DNS related operations
type DNSService interface {
	FindOrCreateZoneIDByName(config *config.Config, name string) (string, error)
	DeleteZoneRecordIfExists(config *config.Config, zoneID string, recordName string) error
	CreateZoneRecord(config *config.Config, zoneID string, record *DNSRecord) error
}

// CreateDNSRecord does the necessary operations to create a DNS record without issues in an cloud provider
func CreateDNSRecord(config *config.Config, aRecordIP string, dnsService DNSService) error {
	domainName := config.RunConfig.DomainName
	if err := isDomainValid(domainName); err != nil {
		return err
	}

	domainParts := strings.Split(domainName, ".")

	// example:
	// domainParts := []string{"test","example","com"}
	zoneName := domainParts[len(domainParts)-2]                 // example
	dnsName := zoneName + "." + domainParts[len(domainParts)-1] // example.com
	aRecordName := domainName + "."                             // test.example.com

	zoneID, err := dnsService.FindOrCreateZoneIDByName(config, dnsName)
	if err != nil {
		return err
	}

	err = dnsService.DeleteZoneRecordIfExists(config, zoneID, aRecordName)
	if err != nil {
		return err
	}

	record := &DNSRecord{
		Name: aRecordName,
		IP:   aRecordIP,
		Type: "A",
		TTL:  TTLDefault,
	}
	err = dnsService.CreateZoneRecord(config, zoneID, record)
	if err != nil {
		return err
	}

	return nil
}

// Context captures required info for provider operation
type Context struct {
	config *config.Config
	logger *Logger
}

// Config returns context configuration
func (c Context) Config() *config.Config {
	return c.config
}

// Logger returns logger
func (c Context) Logger() *Logger {
	return c.logger
}

// NewContext Create a new context for the given provider
// valid providers are "gcp", "aws" and "onprem"
func NewContext(c *config.Config) *Context {

	logger := NewLogger(os.Stdout)

	if c.RunConfig.ShowDebug {
		logger.SetDebug(true)
		logger.SetError(true)
		logger.SetWarn(true)
		logger.SetInfo(true)
	}

	if c.RunConfig.ShowWarnings {
		logger.SetWarn(true)
	}

	if c.RunConfig.ShowErrors {
		logger.SetError(true)
	}

	if c.RunConfig.Verbose {
		logger.SetInfo(true)
	}

	return &Context{
		config: c,
		logger: logger,
	}
}
