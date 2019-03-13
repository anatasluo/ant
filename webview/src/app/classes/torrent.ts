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
  Percentage:           number;
  TypeImg:              string;
  Files:				FileInfo[];
}
