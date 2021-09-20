package firstpartydata

import (
	"encoding/json"
	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/util/jsonutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
)

func TestGetGlobalFPDData(t *testing.T) {

	testCases := []struct {
		description     string
		input           []byte
		output          []byte
		expectedFpdData map[string][]byte
		errorExpected   bool
		errorContains   string
	}{
		{
			description: "Site, app and user data present",
			input: []byte(`{
  				"id": "bid_id",
  				"site": {
  				  "id":"reqSiteId",
  				  "page": "http://www.foobar.com/1234.html",
  				  "publisher": {
  				    "id": "1"
  				  },
				  "ext":{
				   "data": {"somesitefpd": "sitefpdDataTest"}
				  }
  				},
  				"user": {
  				  "id": "reqUserID",
  				  "yob": 1982,
  				  "gender": "M",
				  "ext":{
				  	"data": {"someuserfpd": "userfpdDataTest"}
				  }
  				},
  				"app": {
  				  "id": "appId",
  				  "data": 123,
  				  "ext": {
				     "data": {"someappfpd": "appfpdDataTest"}
				  }
  				},
  				"tmax": 5000,
  				"source": {
  				  "tid": "ad839de0-5ae6-40bb-92b2-af8bad6439b3"
  				}
			}`),
			output: []byte(`{
  				"id": "bid_id",
  				"site": {
  				  "id":"reqSiteId",
  				  "page": "http://www.foobar.com/1234.html",
  				  "publisher": {
  				    "id": "1"
  				  },
				  "ext": {
				  }
  				},
  				"user": {
  				  "id": "reqUserID",
  				  "yob": 1982,
  				  "gender": "M",
				  "ext": {
				  }
  				},
  				"app": {
  				  "id": "appId",
  				  "data": 123,
				  "ext": {
				  }
  				},
  				"tmax": 5000,
  				"source": {
  				  "tid": "ad839de0-5ae6-40bb-92b2-af8bad6439b3"
  				}
			}`),
			expectedFpdData: map[string][]byte{
				"site": []byte(`{"somesitefpd": "sitefpdDataTest"}`),
				"user": []byte(`{"someuserfpd": "userfpdDataTest"}`),
				"app":  []byte(`{"someappfpd": "appfpdDataTest"}`),
			},
			errorExpected: false,
			errorContains: "",
		},
		{
			description: "App FPD only present",
			input: []byte(`{
  				"id": "bid_id",
  				"site": {
  				  "id":"reqSiteId",
  				  "page": "http://www.foobar.com/1234.html",
  				  "publisher": {
  				    "id": "1"
  				  }
  				},
  				"app": {
  				  "id": "appId",
  				  "ext": {
					"data": {"someappfpd": "appfpdDataTest"}
                  }
  				},
  				"tmax": 5000,
  				"source": {
  				  "tid": "ad839de0-5ae6-40bb-92b2-af8bad6439b3"
  				}
			}`),
			output: []byte(`{
  				"id": "bid_id",
  				"site": {
  				  "id":"reqSiteId",
  				  "page": "http://www.foobar.com/1234.html",
  				  "publisher": {
  				    "id": "1"
  				  }
  				},
  				"app": {
  				  "id": "appId",
				  "ext": {}
  				},
  				"tmax": 5000,
  				"source": {
  				  "tid": "ad839de0-5ae6-40bb-92b2-af8bad6439b3"
  				}
			}`),
			expectedFpdData: map[string][]byte{
				"app":  []byte(`{"someappfpd": "appfpdDataTest"}`),
				"user": {},
				"site": {},
			},
			errorExpected: false,
			errorContains: "",
		},
		{
			description: "Site FPD different format",
			input: []byte(`{
  				"id": "bid_id",
  				"site": {
  				  "id":"reqSiteId",
  				  "page": "http://www.foobar.com/1234.html",
  				  "publisher": {
  				    "id": "1"
  				  },
				  "ext": {
					"data": {"someappfpd": true}
                  }
  				},
  				"app": {
  				  "id": "appId"
  				},
  				"tmax": 5000,
  				"source": {
  				  "tid": "ad839de0-5ae6-40bb-92b2-af8bad6439b3"
  				}
			}`),
			output: []byte(`{
  				"id": "bid_id",
  				"site": {
  				  "id":"reqSiteId",
  				  "page": "http://www.foobar.com/1234.html",
  				  "publisher": {
  				    "id": "1"
  				  },
                  "ext": {}
  				},
  				"app": {
  				  "id": "appId"
  				},
  				"tmax": 5000,
  				"source": {
  				  "tid": "ad839de0-5ae6-40bb-92b2-af8bad6439b3"
  				}
			}`),
			expectedFpdData: map[string][]byte{
				"app":  {},
				"user": {},
				"site": []byte(`{"someappfpd": true}`),
			},
			errorExpected: false,
			errorContains: "",
		},
	}
	for _, test := range testCases {
		var inputTestReq openrtb_ext.RequestWrapper
		err := json.Unmarshal(test.input, &inputTestReq)
		assert.NoError(t, err, "Error should be nil")

		fpd, err := ExtractGlobalFPDData(&inputTestReq)
		assert.NoError(t, err, "Error should be nil")

		var outputTestReq openrtb_ext.RequestWrapper
		err = json.Unmarshal(test.output, &outputTestReq)
		assert.NoError(t, err, "Error should be nil")

		if test.errorExpected {
			assert.Error(t, err, "Error should not be nil")
			//result should be still returned
			assert.Equal(t, inputTestReq, outputTestReq, "Result is incorrect")
			assert.True(t, strings.Contains(err.Error(), test.errorContains))
		} else {
			assert.NoError(t, err, "Error should be nil")
			if fpd[userKey] != nil {
				assert.JSONEq(t, string(inputTestReq.User.Ext), string(outputTestReq.User.Ext), "Result is incorrect")
				assert.Equal(t, test.expectedFpdData[userKey], fpd[userKey], "FPD is incorrect")
			}
			if fpd[appKey] != nil {
				assert.JSONEq(t, string(inputTestReq.App.Ext), string(outputTestReq.App.Ext), "Result is incorrect")
				assert.Equal(t, test.expectedFpdData[appKey], fpd[appKey], "FPD is incorrect")
			}
			if fpd[siteKey] != nil {
				assert.JSONEq(t, string(inputTestReq.Site.Ext), string(outputTestReq.Site.Ext), "Result is incorrect")
				assert.Equal(t, test.expectedFpdData[siteKey], fpd[siteKey], "FPD is incorrect")
			}
		}
	}
}

func TestExtractOpenRtbGlobalFPD(t *testing.T) {

	testCases := []struct {
		description     string
		input           []byte
		output          []byte
		expectedFpdData map[string][]openrtb2.Data
	}{
		{
			description: "Site, app and user data present",
			input: []byte(`{
  				"id": "bid_id",
			 	"imp":[{"id":"impid"}],
  				"site": {
  				  "id":"reqSiteId",
				  "content": {
					"data":[
						{ 
						  "id": "siteDataId1",
						  "name": "siteDataName1"
						},
						{
 						  "id": "siteDataId2",
            			  "name": "siteDataName2"
						}
					]
				  }
  				},
  				"user": {
  				  "id": "reqUserID",
  				  "yob": 1982,
  				  "gender": "M",
				  "data":[
						{ 
						  "id": "userDataId1",
						  "name": "userDataName1"
						}
					]
  				},
  				"app": {
  				  "id": "appId",
					"content":{
						"data": [
							{ 
							  "id": "appDataId1",
							  "name": "appDataName1"
							}
						]
					}
  				}
			}`),
			output: []byte(`{
  				"id": "bid_id",
				"imp":[{"id":"impid"}],
  				"site": {
  				  "id":"reqSiteId",
				  "content": {}
  				},
  				"user": {
  				  "id": "reqUserID",
  				  "yob": 1982,
  				  "gender": "M"
  				},
  				"app": {
  				  "id": "appId",
				  "content": {}
  				}
			}`),
			expectedFpdData: map[string][]openrtb2.Data{
				siteContentDataKey: {{ID: "siteDataId1", Name: "siteDataName1"}, {ID: "siteDataId2", Name: "siteDataName2"}},
				userDataKey:        {{ID: "userDataId1", Name: "userDataName1"}},
				appContentDataKey:  {{ID: "appDataId1", Name: "appDataName1"}},
			},
		}, {
			description: "User only data present",
			input: []byte(`{
  				"id": "bid_id",
			 	"imp":[{"id":"impid"}],
  				"site": {
  				  "id":"reqSiteId"
  				},
  				"user": {
  				  "id": "reqUserID",
  				  "yob": 1982,
  				  "gender": "M",
				  "data":[
						{ 
						  "id": "userDataId1",
						  "name": "userDataName1"
						}
					]
  				},
  				"app": {
  				  "id": "appId"
  				}
			}`),
			output: []byte(`{
  				"id": "bid_id",
				"imp":[{"id":"impid"}],
  				"site": {
  				  "id":"reqSiteId"
  				},
  				"user": {
  				  "id": "reqUserID",
  				  "yob": 1982,
  				  "gender": "M"
  				},
  				"app": {
  				  "id": "appId"
  				}
			}`),
			expectedFpdData: map[string][]openrtb2.Data{
				siteContentDataKey: nil,
				userDataKey:        {{ID: "userDataId1", Name: "userDataName1"}},
				appContentDataKey:  nil,
			},
		},
	}
	for _, test := range testCases {

		var req openrtb2.BidRequest
		err := json.Unmarshal(test.input, &req)
		assert.NoError(t, err, "Error should be nil")

		res := ExtractOpenRtbGlobalFPD(&req)

		resReq, err := json.Marshal(req)
		assert.NoError(t, err, "Error should be nil")

		assert.JSONEq(t, string(test.output), string(resReq), "Result request is incorrect")
		assert.Equal(t, test.expectedFpdData[siteContentDataKey], res[siteContentDataKey], "siteContentData data is incorrect")
		assert.Equal(t, test.expectedFpdData[userDataKey], res[userDataKey], "userData is incorrect")
		assert.Equal(t, test.expectedFpdData[appContentDataKey], res[appContentDataKey], "appContentData is incorrect")

	}
}

func TestPreprocessFPD(t *testing.T) {

	if specFiles, err := ioutil.ReadDir("./tests/preprocessfpd"); err == nil {
		for _, specFile := range specFiles {
			fileName := "./tests/preprocessfpd/" + specFile.Name()

			fpdFile, err := loadFpdFile(fileName)
			if err != nil {
				t.Errorf("Unable to load file: %s", fileName)
			}
			var extReq openrtb_ext.ExtRequestPrebid
			err = json.Unmarshal(fpdFile.InputRequestData, &extReq)
			if err != nil {
				t.Errorf("Unable to unmarshal input request: %s", fileName)
			}
			fpdData, reqExtPrebid := ExtractBidderConfigFPD(extReq)

			assert.Nil(t, reqExtPrebid.BidderConfigs, "Bidder specific FPD config should be removed from request")

			assert.Equal(t, len(fpdFile.BiddersFPD), len(fpdData), "Incorrect fpd data")

			for k, v := range fpdFile.BiddersFPD {

				if v.Site != nil {
					tempSiteExt := fpdData[k].Site.Ext
					jsonutil.DiffJson(t, "site.ext is incorrect", v.Site.Ext, tempSiteExt)
					//compare extensions first and the site objects without extensions
					//in case two or more bidders share same config(pointer), ext should be returned back
					v.Site.Ext = nil
					fpdData[k].Site.Ext = nil
					assert.Equal(t, v.Site, fpdData[k].Site, "Incorrect site fpd data")
					fpdData[k].Site.Ext = tempSiteExt
				}

				if v.App != nil {

					tempAppExt := fpdData[k].App.Ext
					jsonutil.DiffJson(t, "app.ext is incorrect", v.App.Ext, tempAppExt)
					//compare extensions first and the app objects without extensions
					v.App.Ext = nil
					fpdData[k].App.Ext = nil
					assert.Equal(t, v.App, fpdData[k].App, "Incorrect app fpd data")
					fpdData[k].App.Ext = tempAppExt
				}

				if v.User != nil {
					tempUserExt := fpdData[k].User.Ext
					jsonutil.DiffJson(t, "user.ext is incorrect", v.User.Ext, tempUserExt)
					//compare extensions first and the user objects without extensions
					v.User.Ext = nil
					fpdData[k].User.Ext = nil
					assert.Equal(t, v.User, fpdData[k].User, "Incorrect user fpd data")
					fpdData[k].User.Ext = tempUserExt
				}

			}
		}
	}
}

func TestBuildResolvedFPDForBidders(t *testing.T) {

	if specFiles, err := ioutil.ReadDir("./tests/applyfpd"); err == nil {
		for _, specFile := range specFiles {
			fileName := "./tests/applyfpd/" + specFile.Name()

			fpdFile, err := loadFpdFile(fileName)
			if err != nil {
				t.Errorf("Unable to load file: %s", fileName)
			}

			var inputReq openrtb2.BidRequest
			err = json.Unmarshal(fpdFile.InputRequestData, &inputReq)
			if err != nil {
				t.Errorf("Unable to unmarshal input request: %s", fileName)
			}

			var inputReqCopy openrtb2.BidRequest
			err = json.Unmarshal(fpdFile.InputRequestData, &inputReqCopy)
			if err != nil {
				t.Errorf("Unable to unmarshal input request: %s", fileName)
			}

			var outputReq openrtb2.BidRequest
			err = json.Unmarshal(fpdFile.OutputRequestData, &outputReq)
			if err != nil {
				t.Errorf("Unable to unmarshal output request: %s", fileName)
			}

			reqExtFPD := make(map[string][]byte, 3)
			reqExtFPD["site"] = fpdFile.FirstPartyData["site"]
			reqExtFPD["app"] = fpdFile.FirstPartyData["app"]
			reqExtFPD["user"] = fpdFile.FirstPartyData["user"]

			reqFPD := make(map[string][]openrtb2.Data, 3)

			reqFPDSiteContentData := fpdFile.FirstPartyData[siteContentDataKey]
			if len(reqFPDSiteContentData) > 0 {
				var siteConData []openrtb2.Data
				err = json.Unmarshal(reqFPDSiteContentData, &siteConData)
				if err != nil {
					t.Errorf("Unable to unmarshal site.content.data: %s", fileName)
				}
				reqFPD[siteContentDataKey] = siteConData
			}

			reqFPDAppContentData := fpdFile.FirstPartyData[appContentDataKey]
			if len(reqFPDAppContentData) > 0 {
				var appConData []openrtb2.Data
				err = json.Unmarshal(reqFPDAppContentData, &appConData)
				if err != nil {
					t.Errorf("Unable to unmarshal app.content.data: %s", fileName)
				}
				reqFPD[appContentDataKey] = appConData
			}

			reqFPDUserData := fpdFile.FirstPartyData[userDataKey]
			if len(reqFPDUserData) > 0 {
				var userData []openrtb2.Data
				err = json.Unmarshal(reqFPDUserData, &userData)
				if err != nil {
					t.Errorf("Unable to unmarshal app.content.data: %s", fileName)
				}
				reqFPD[userDataKey] = userData
			}
			if fpdFile.BiddersFPD == nil {
				fpdFile.BiddersFPD = make(map[openrtb_ext.BidderName]*openrtb_ext.ORTB2)
				fpdFile.BiddersFPD["appnexus"] = &openrtb_ext.ORTB2{}
			}

			resultFPD, errL := ResolveFPDData(&inputReq, fpdFile.BiddersFPD, reqExtFPD, reqFPD, []string{"appnexus"})

			if len(errL) == 0 {
				assert.Equal(t, inputReq, inputReqCopy, "Original request should not be modified")

				bidderFPD := resultFPD["appnexus"]

				if bidderFPD.Site != nil && len(bidderFPD.Site.Ext) > 0 {
					resSiteExt := bidderFPD.Site.Ext
					expectedSiteExt := outputReq.Site.Ext
					bidderFPD.Site.Ext = nil
					outputReq.Site.Ext = nil
					jsonutil.DiffJson(t, "site.ext is incorrect", resSiteExt, expectedSiteExt)
				}
				if bidderFPD.App != nil && len(bidderFPD.App.Ext) > 0 {
					resAppExt := bidderFPD.App.Ext
					expectedAppExt := outputReq.App.Ext
					bidderFPD.App.Ext = nil
					outputReq.App.Ext = nil
					jsonutil.DiffJson(t, "app.ext is incorrect", resAppExt, expectedAppExt)
				}
				if bidderFPD.User != nil && len(bidderFPD.User.Ext) > 0 {
					resUserExt := bidderFPD.User.Ext
					expectedUserExt := outputReq.User.Ext
					bidderFPD.User.Ext = nil
					outputReq.User.Ext = nil
					jsonutil.DiffJson(t, "user.ext is incorrect", resUserExt, expectedUserExt)
				}
			} else {
				for i := range fpdFile.ValidationErrors {
					assert.Contains(t, errL[i].Error(), fpdFile.ValidationErrors[i], "Incorrect first party data warning message")
				}
			}

		}
	}
}

func loadFpdFile(filename string) (fpdFile, error) {
	var fileData fpdFile
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return fileData, err
	}
	err = json.Unmarshal(fileContents, &fileData)
	if err != nil {
		return fileData, err
	}

	return fileData, nil
}

type fpdFile struct {
	InputRequestData  json.RawMessage                               `json:"inputRequestData,omitempty"`
	OutputRequestData json.RawMessage                               `json:"outputRequestData,omitempty"`
	BiddersFPD        map[openrtb_ext.BidderName]*openrtb_ext.ORTB2 `json:"biddersFPD,omitempty"`
	FirstPartyData    map[string]json.RawMessage                    `json:"firstPartyData,omitempty"`
	ValidationErrors  []string                                      `json:"validationErrors,omitempty"`
}
