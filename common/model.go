package common

type User struct {
	ID          string
	ChannelType string
	Name        string
	Email       string
	Phone       string
}

type Location struct {
	Name string
	Lat  float32
	Lng  float32
}

type RapidMessage struct {
	ID          string
	App         string
	Type        string
	ChannelType string
	ChannelId   string
	Text        string
	Url         string
	Location    *Location
	Sender      User
	Recipient   User
	Extra       map[string]interface{}
}
