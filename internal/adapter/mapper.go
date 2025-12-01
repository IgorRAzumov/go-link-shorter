package adapter

import (
	"github.com/IgorRAzumov/go-link-shorter/internal/domain/model"
)

func FromDomainLink(domainLink *model.Link) *Link {
	return &Link{
		ShortKey: domainLink.ShortKey,
		FullURL:  domainLink.FullURL,
	}
}
