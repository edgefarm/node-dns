package feed

// Feed is an interface to enable different sources to obtain of host/ip entries
type FeedIf interface {
	Update() error
	GetDnsMap() map[string]string
}

type Feed struct {
	// FeedDnsMap is a map of hostnames with their corresponding IP addresses
	FeedDnsMap map[string]string
}
