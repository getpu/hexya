// Copyright 2017 NDP Systèmes. All Rights Reserved.
// See LICENSE file for full licensing details.

package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/hexya-erp/hexya/hexya/models/fieldtype"
	"github.com/hexya-erp/hexya/hexya/models/security"
	"github.com/hexya-erp/hexya/hexya/models/types/dates"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBaseModelMethods(t *testing.T) {
	Convey("Testing base model methods", t, func() {
		SimulateInNewEnvironment(security.SuperUserID, func(env Environment) {
			userModel := Registry.MustGet("User")
			userJane := userModel.Search(env, userModel.Field("Email").Equals("jane.smith@example.com"))
			Convey("LastUpdate", func() {
				So(userJane.Get("LastUpdate").(dates.DateTime).Sub(userJane.Get("WriteDate").(dates.DateTime).Time), ShouldBeLessThanOrEqualTo, 1*time.Second)
				newUser := userModel.Create(env, FieldMap{
					"Name":    "Alex Smith",
					"Email":   "jsmith@example.com",
					"IsStaff": true,
					"Nums":    1,
				})
				time.Sleep(1*time.Second + 100*time.Millisecond)
				So(newUser.Get("WriteDate").(dates.DateTime).IsZero(), ShouldBeTrue)
				So(newUser.Get("LastUpdate").(dates.DateTime).Sub(newUser.Get("CreateDate").(dates.DateTime).Time), ShouldBeLessThanOrEqualTo, 1*time.Second)
			})
			Convey("Load and Read", func() {
				userJane = userJane.Call("Load", []string{"ID", "Name", "Age", "Posts", "Profile"}).(RecordSet).Collection()
				res := userJane.Call("Read", []string{"Name", "Age", "Posts", "Profile"})
				So(res, ShouldHaveLength, 1)
				fMap := res.([]FieldMap)[0]
				So(fMap, ShouldHaveLength, 5)
				So(fMap, ShouldContainKey, "Name")
				So(fMap["Name"], ShouldEqual, "Jane A. Smith")
				So(fMap, ShouldContainKey, "Age")
				So(fMap["Age"], ShouldEqual, 24)
				So(fMap, ShouldContainKey, "Posts")
				So(fMap["Posts"].(RecordSet).Collection().Ids(), ShouldHaveLength, 2)
				So(fMap, ShouldContainKey, "Profile")
				So(fMap["Profile"].(RecordSet).Collection().Get("ID"), ShouldEqual, userJane.Get("Profile").(RecordSet).Collection().Get("ID"))
				So(fMap, ShouldContainKey, "id")
				So(fMap["id"], ShouldEqual, userJane.Ids()[0])
			})
			Convey("Copy", func() {
				userJane.Call("Write", FieldMap{"Password": "Jane's Password"})
				userJaneCopy := userJane.Call("Copy", FieldMap{"Name": "Jane's Copy", "Email2": "js@example.com"}).(RecordSet).Collection()
				So(userJaneCopy.Get("Name"), ShouldEqual, "Jane's Copy")
				So(userJaneCopy.Get("Email"), ShouldEqual, "jane.smith@example.com")
				So(userJaneCopy.Get("Email2"), ShouldEqual, "js@example.com")
				So(userJaneCopy.Get("Password"), ShouldBeBlank)
				So(userJaneCopy.Get("Age"), ShouldEqual, 24)
				So(userJaneCopy.Get("Nums"), ShouldEqual, 2)
				So(userJaneCopy.Get("Posts").(RecordSet).Collection().Len(), ShouldEqual, 0)
			})
			Convey("FieldGet and FieldsGet", func() {
				fInfo := userJane.Call("FieldGet", FieldName("Name")).(*FieldInfo)
				So(fInfo.String, ShouldEqual, "Name")
				So(fInfo.Help, ShouldEqual, "The user's username")
				So(fInfo.Type, ShouldEqual, fieldtype.Char)
				fInfos := userJane.Call("FieldsGet", FieldsGetArgs{}).(map[string]*FieldInfo)
				So(fInfos, ShouldHaveLength, 30)
			})
			Convey("NameGet", func() {
				So(userJane.Get("DisplayName"), ShouldEqual, "Jane A. Smith")
				profile := userJane.Get("Profile").(RecordSet).Collection()
				So(profile.Get("DisplayName"), ShouldEqual, fmt.Sprintf("Profile(%d)", profile.Get("ID")))
			})
			Convey("DefaultGet", func() {
				defaults := userJane.Call("DefaultGet").(FieldMap)
				So(defaults, ShouldHaveLength, 2)
				So(defaults, ShouldContainKey, "status_json")
				So(defaults["status_json"], ShouldEqual, 12)
				So(defaults, ShouldContainKey, "hexya_external_id")
			})
			Convey("Onchange", func() {
				res := userJane.Call("Onchange", OnchangeParams{
					Fields:   []string{"Name"},
					Onchange: map[string]string{"Name": "1"},
					Values:   FieldMap{"Name": "William", "Email": "will@example.com"},
				}).(OnchangeResult)
				fMap := res.Value.FieldMap()
				So(fMap, ShouldHaveLength, 1)
				So(fMap, ShouldContainKey, "decorated_name")
				So(fMap["decorated_name"], ShouldEqual, "User: William [<will@example.com>]")
			})
			Convey("CheckRecursion", func() {
				So(userJane.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag1 := env.Pool("Tag").Call("Create", FieldMap{
					"Name": "Tag1",
				}).(RecordSet).Collection()
				So(tag1.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag2 := env.Pool("Tag").Call("Create", FieldMap{
					"Name":   "Tag2",
					"Parent": tag1,
				}).(RecordSet).Collection()
				So(tag2.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag3 := env.Pool("Tag").Call("Create", FieldMap{
					"Name":   "Tag1",
					"Parent": tag2,
				}).(RecordSet).Collection()
				So(tag3.Call("CheckRecursion").(bool), ShouldBeTrue)
				tag1.Set("Parent", tag3)
				So(tag1.Call("CheckRecursion").(bool), ShouldBeFalse)
				So(tag2.Call("CheckRecursion").(bool), ShouldBeFalse)
				So(tag3.Call("CheckRecursion").(bool), ShouldBeFalse)
			})
			Convey("Browse", func() {
				browsedUser := env.Pool("User").Call("Browse", []int64{userJane.Ids()[0]}).(RecordSet).Collection()
				So(browsedUser.Ids(), ShouldHaveLength, 1)
				So(browsedUser.Ids(), ShouldContain, userJane.Ids()[0])
			})
			Convey("Equals", func() {
				browsedUser := env.Pool("User").Call("Browse", []int64{userJane.Ids()[0]}).(RecordSet).Collection()
				So(browsedUser.Call("Equals", userJane), ShouldBeTrue)
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field("Name").Equals("John Smith")).(RecordSet).Collection()
				So(userJohn.Call("Equals", userJane), ShouldBeFalse)
				johnAndJane := userJohn.Union(userJane)
				usersJ := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field("Name").Like("J% Smith")).(RecordSet).Collection()
				So(usersJ.Records(), ShouldHaveLength, 2)
				So(usersJ.Equals(johnAndJane), ShouldBeTrue)
			})
			Convey("Subtract", func() {
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field("Name").Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				So(johnAndJane.Subtract(userJane).Equals(userJohn), ShouldBeTrue)
				So(johnAndJane.Subtract(userJohn).Equals(userJane), ShouldBeTrue)
			})
			Convey("Intersect", func() {
				userJohn := env.Pool("User").Call("Search", env.Pool("User").Model().
					Field("Name").Equals("John Smith")).(RecordSet).Collection()
				johnAndJane := userJohn.Union(userJane)
				So(johnAndJane.Intersect(userJane).Equals(userJane), ShouldBeTrue)
				So(johnAndJane.Call("Intersect", userJohn).(RecordSet).Collection().Equals(userJohn), ShouldBeTrue)
			})
			Convey("ConvertLimitToInt", func() {
				So(ConvertLimitToInt(12), ShouldEqual, 12)
				So(ConvertLimitToInt(false), ShouldEqual, -1)
				So(ConvertLimitToInt(0), ShouldEqual, 0)
				So(ConvertLimitToInt(nil), ShouldEqual, 80)
			})
			Convey("CartesianProduct", func() {
				tagA := env.Pool("Tag").Call("Create", FieldMap{"Name": "A"}).(RecordSet).Collection()
				tagB := env.Pool("Tag").Call("Create", FieldMap{"Name": "B"}).(RecordSet).Collection()
				tagC := env.Pool("Tag").Call("Create", FieldMap{"Name": "C"}).(RecordSet).Collection()
				tagD := env.Pool("Tag").Call("Create", FieldMap{"Name": "D"}).(RecordSet).Collection()
				tagE := env.Pool("Tag").Call("Create", FieldMap{"Name": "E"}).(RecordSet).Collection()
				tagF := env.Pool("Tag").Call("Create", FieldMap{"Name": "F"}).(RecordSet).Collection()
				tagG := env.Pool("Tag").Call("Create", FieldMap{"Name": "G"}).(RecordSet).Collection()
				tagsAB := tagA.Union(tagB)
				tagsCD := tagC.Union(tagD)
				tagsEFG := tagE.Union(tagF).Union(tagG)

				contains := func(product []*RecordCollection, collections ...*RecordCollection) bool {
				productLoop:
					for _, p := range product {
						for _, c := range collections {
							if c.Equals(p) {
								break productLoop
							}
						}
						return false
					}
					return true
				}

				product1 := tagsAB.CartesianProduct(tagsCD)
				So(product1, ShouldHaveLength, 4)
				So(contains(product1,
					tagA.Union(tagC),
					tagA.Union(tagD),
					tagB.Union(tagC),
					tagB.Union(tagD)), ShouldBeTrue)

				product2 := tagsAB.CartesianProduct(tagsEFG)
				So(product2, ShouldHaveLength, 6)
				So(contains(product2,
					tagA.Union(tagE),
					tagA.Union(tagF),
					tagA.Union(tagG),
					tagB.Union(tagE),
					tagB.Union(tagF),
					tagB.Union(tagG)), ShouldBeTrue)

				product3 := tagsAB.CartesianProduct(tagsCD, tagsEFG)
				So(product3, ShouldHaveLength, 12)
				So(contains(product3,
					tagA.Union(tagC).Union(tagE),
					tagA.Union(tagC).Union(tagF),
					tagA.Union(tagC).Union(tagG),
					tagA.Union(tagD).Union(tagE),
					tagA.Union(tagD).Union(tagF),
					tagA.Union(tagD).Union(tagG),
					tagB.Union(tagC).Union(tagE),
					tagB.Union(tagC).Union(tagF),
					tagB.Union(tagC).Union(tagG),
					tagB.Union(tagD).Union(tagE),
					tagB.Union(tagD).Union(tagF),
					tagB.Union(tagD).Union(tagG)), ShouldBeTrue)
			})
		})
	})
}
