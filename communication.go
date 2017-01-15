package cheesegull

// In this file there are services related to the communication between API and
// mirror.

// CommunicationService is a service able to send data to another end of the
// application and to provide methods of receiving the data on the other side.
type CommunicationService interface {
	SendBeatmapRequest(int) error
	BeatmapRequestsChannel() (<-chan int, error)
}
