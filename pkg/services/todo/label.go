package todo

type Label struct {
	ID     uint   `gorm:"primary_key"`
	Color  string `gorm:"color"`
	Text   string `gorm:"text"`
	TodoId uint
}

func (l Label) TableName() string {
	return "labels"
}
