package lib

import "strings"

const INERTRA_LEN = 6
const FLUSSONIC_LEN = 9

// шаблон
// `/streaming/muztv/324/vl2w/segment-1597220444-01972046.ts`
func isInetraTranscoder(len int) bool {
	return len == INERTRA_LEN
}

// шаблон
// `/domashniy/tracks-v1a1/2020/08/13/11/38/56-06000.ts`
func isFlussonicTranscoder(len int) bool {
	return len == FLUSSONIC_LEN
}

func isFlussonicPlaylist(len int) bool {
	return len == 4
}

func isFlussonicMultibitrate(len int) bool {
	return len == 3
}

func isInetraMultibitrate(len int) bool {
	return len == 5
}

func isPlaylist(index string) bool {
	return strings.Contains(index, ".m3u8")
}

func isMediaFile(index string) bool {
	return strings.Contains(index, ".ts")
}

func (l Log) IsPlaylistStream() bool {
	return isPlaylist(l._splitUri.index)
}

func (l Log) IsMediaStream() bool {
	return isMediaFile(l._splitUri.index)
}
