package model

type PageTypeEnum int

const (
	PageTypeDefault PageTypeEnum = iota // 0
	PageTypeLanding                     // 1
	PageTypeCatalog                     // 2
	PageTypeArticle                     // 3
	PageTypeContact                     // 4
	PageTypeHome                        // 5
)

func (t PageTypeEnum) String() string {
	switch t {
	case PageTypeDefault:
		return "default"
	case PageTypeLanding:
		return "landing"
	case PageTypeCatalog:
		return "catalog"
	case PageTypeArticle:
		return "article"
	case PageTypeContact:
		return "contact"
	case PageTypeHome:
		return "home"
	default:
		return "unknown"
	}
}

type PageStatusEnum int

const (
	PageStatusDraft     PageStatusEnum = iota // 0
	PageStatusPublished                       // 1
	PageStatusArchived                        // 2
	PageStatusDeleted                         // 3
)

func (s PageStatusEnum) String() string {
	switch s {
	case PageStatusDraft:
		return "draft"
	case PageStatusPublished:
		return "published"
	case PageStatusArchived:
		return "archived"
	case PageStatusDeleted:
		return "deleted"
	default:
		return "unknown"
	}
}

type BlockTypeEnum int

const (
	BlockContentText  BlockTypeEnum = iota // 0
	BlockContentImage                      // 1
	BlockContentVideo                      // 2
	BlockContentLink                       // 3
	BlockContentHTML                       // 4
)

func (c BlockTypeEnum) String() string {
	switch c {
	case BlockContentText:
		return "text"
	case BlockContentImage:
		return "image"
	case BlockContentVideo:
		return "video"
	case BlockContentLink:
		return "link"
	case BlockContentHTML:
		return "html"
	default:
		return "unknown"
	}
}
