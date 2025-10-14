package objectstorage

type IObjectStorage interface {

	/*
		UploadFile 上传文件
		key: 对象存储上的文件路径+文件名
		filePath: 本地文件路径+文件名
	*/
	UploadFile(key string, filePath string) error

	/*
		DownloadFile 下载文件
		key: 对象存储上的文件路径+文件名
		filePath: 本地文件路径+文件名
	*/
	DownloadFile(key string, filePath string) error

	/*
		IsFileExist 判断文件是否存在
		key: 对象存储上的文件路径+文件名
	*/
	IsFileExist(key string) (bool, error)
}
