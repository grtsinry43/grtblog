package album

import "errors"

var (
	ErrAlbumNotFound       = errors.New("相册不存在")
	ErrAlbumShortURLExists = errors.New("相册短链接已存在")
	ErrPhotoNotFound       = errors.New("照片不存在")
)
