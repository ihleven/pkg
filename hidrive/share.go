package hidrive

// func (c *HiDriveClient) GetShare(path string) (interface{}, error) {

// 	params := url.Values{
// 		// "path":   {path},
// 		"fields": {"count,created,file_type,has_password,is_encrypted,id,last_modified,maxcount,name,password,path,pid,readable,remaining,share_type,size,status,ttl,uri,valid_until,writable"},
// 	}
// 	var result interface{}

// 	err := c.Get("/share", params, &result)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return result, nil
// }

// func (c *HiDriveClient) GetShareToken(id string, password string) (interface{}, error) {
// 	type ShareToken struct {
// 		ExpiresIn   int    `json:"expires_in"`   // 14400
// 		AccessToken string `json:"access_token"` // "L2oJ32O84QhWdKhk1Jv8"
// 		TokenType   string `json:"token_type"`   // "Bearer"
// 	}

// 	params := url.Values{
// 		"id":       {id},
// 		"password": {password},
// 	}
// 	var result ShareToken

// 	err := c.BaseExec("POST", "https://api.hidrive.strato.com/2.1/share/token?id="+params.Encode(), strings.NewReader(params.Encode()), &result)
// 	fmt.Println(err, result)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &result, nil
// }
