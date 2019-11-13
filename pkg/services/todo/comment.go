package todo

type Comment struct {
	ID     uint   `gorm:"primary_key"`
	Text   string `gorm:"text"`
	TodoId uint
}

func (c Comment) TableName() string {
	return "comments"
}
