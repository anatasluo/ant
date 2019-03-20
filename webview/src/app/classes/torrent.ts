class FileInfo {
  Path:				string;
  Priority:			number;
  Size:				string;
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
  logUpdateTime:        number;
  leftTime:             string;
  downloadSpeed:        string;
}
