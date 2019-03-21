package engine

import (
	"github.com/anacrolix/torrent"
	"io"
)


// SeekableContent describes an io.ReadSeeker that can be closed as well.
type SeekableContent interface {
	io.ReadSeeker
	io.Closer
}

// FileEntry helps reading a torrent file.
type FileEntry struct {
	*torrent.File
	torrent.Reader
}

// Seek seeks to the correct file position, paying attention to the offset.
func (f *FileEntry) Seek(offset int64, whence int) (int64, error) {
	return f.Reader.Seek(offset+f.File.Offset(), whence)
}

func getLargestFile(singleTorrent *torrent.Torrent) *torrent.File {
	var target *torrent.File
	var maxSize int64

	for _, file := range singleTorrent.Files() {
		if maxSize < file.Length() {
			maxSize = file.Length()
			target = file
		}
	}

	firstPieceIndex := target.Offset() * int64(singleTorrent.NumPieces()) / singleTorrent.Length()
	endPieceIndex := (target.Offset() + target.Length()) * int64(singleTorrent.NumPieces()) / singleTorrent.Length()
	for idx := firstPieceIndex; idx <= endPieceIndex*10/100; idx++ {
		singleTorrent.Piece(int(idx)).SetPriority(torrent.PiecePriorityNow)
	}
	return target
}

//TODO
func (engine *Engine)GetReaderFromTorrent(singleTorrent *torrent.Torrent, fileID string)(SeekableContent, *torrent.File, error)  {
	return getReaderFromFile(getLargestFile(singleTorrent))
}

func getReaderFromFile(singleFile *torrent.File)(SeekableContent, *torrent.File, error) {
	singleTorrent := singleFile.Torrent()
	fileReader := singleTorrent.NewReader()

	//Read ahead 1% of the file
	fileReader.SetReadahead(singleFile.Length() / 100)
	_, err := fileReader.Seek(singleFile.Offset(), io.SeekStart)

	return &FileEntry{
		File: singleFile,
		Reader: fileReader,
	}, singleFile, err
}
