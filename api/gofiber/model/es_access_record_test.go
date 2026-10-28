package model

//func TestAccessRecord_CreateElasticTypedDocument(t *testing.T) {
//
//	myHF.InitDefaultTypedElasticSearch()
//
//	tests := []struct {
//		name   string
//		fields AccessRecord
//		//want    *index.Response
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//		{"t0", AccessRecord{"tester001", "fake_device_id", "dname", []string{"1.2.3.4"}, time.Now()}, false},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := tt.fields.CreateElasticTypedDocument()
//			if (err != nil) != tt.wantErr {
//				t.Errorf("CreateElasticTypedDocument() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//
//			t.Logf("%v", got)
//			//if !reflect.DeepEqual(got, tt.want) {
//			//	t.Errorf("CreateElasticTypedDocument() got = %v, want %v", got, tt.want)
//			//}
//		})
//	}
//}

// func TestAccessRecord_ListAll(t *testing.T) {
// 	_, _ = elastic.InitDefaultTypedElasticSearch()
// 	tests := []struct {
// 		name   string
// 		fields AccessRecord
// 		//want    *search.Response
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 		{"t0", AccessRecord{}, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			got, err := tt.fields.ListAll()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ListAll() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			t.Logf("%v %#v", len(got), got)
// 			//if !reflect.DeepEqual(got, tt.want) {
// 			//	t.Errorf("ListAll() got = %v, want %v", got, tt.want)
// 			//}
// 		})
// 	}
// }
