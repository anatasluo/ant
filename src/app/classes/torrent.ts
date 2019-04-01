class FileInfo {
  Path:				string;
  Priority:			number;
  Size:				string;
}

class TorrentStatusInfo {
  ActivePeers:          number;
  ConnectedSeeders:     number;
  HalfOpenPeers:        number;
  PendingPeers:         number;
  TotalPeers:           number;
}

export class Torrent {
  TorrentName:			string;
  TotalLength:			string;
  Status:				string;
  StoragePath: 		    string;
  HexString:            string;
  Percentage:           string;
  TypeImg:              string;
  Files:				FileInfo[];
  Interval:             number;
  LeftTime:             string;
  DownloadSpeed:        string;
  TorrentStatus:        TorrentStatusInfo;
  UpdateTime:           any;
  StreamURL:            string;
}
