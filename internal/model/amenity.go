package model

type Amenity struct {
	ID        int    `json:"id"`
	NameRU    string `json:"name_ru"`
	NameEN    string `json:"name_en"`
	NameDE    string `json:"name_de"`
	NameTR    string `json:"name_tr"`
	Icon      string `json:"icon"`
	CreatedAt string `json:"created_at"`
}
