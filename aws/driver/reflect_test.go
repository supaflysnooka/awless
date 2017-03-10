/*
Copyright 2017 WALLIX

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestSetFieldsOnAwsStruct(t *testing.T) {
	awsparams := &ec2.RunInstancesInput{}

	err := setFieldWithType("ami", awsparams, "ImageId", awsstr)
	if err != nil {
		t.Fatal(err)
	}
	err = setFieldWithType("t2.micro", awsparams, "InstanceType", awsstr)
	if err != nil {
		t.Fatal(err)
	}
	err = setFieldWithType("5", awsparams, "MaxCount", awsint64)
	if err != nil {
		t.Fatal(err)
	}
	err = setFieldWithType(3, awsparams, "MinCount", awsint64)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := aws.StringValue(awsparams.ImageId), "ami"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	if got, want := aws.StringValue(awsparams.InstanceType), "t2.micro"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	if got, want := aws.Int64Value(awsparams.MaxCount), int64(5); got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
	if got, want := aws.Int64Value(awsparams.MinCount), int64(3); got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

func TestSetFieldWithMultiType(t *testing.T) {
	any := struct {
		Field                       string
		IntField                    int
		BoolPointerField            *bool
		BoolField                   bool
		StringArrayField            []*string
		Int64ArrayField             []*int64
		BooleanValueField           *ec2.AttributeBooleanValue
		StringValueField            *ec2.AttributeValue
		StructAttribute             struct{ Str *string }
		SliceStructPointerAttribute []*struct{ Str1, Str2 *string }
		MapAttribute                map[string]*string
		EmptyMapAttribute           map[string]*string
	}{Field: "initial", MapAttribute: map[string]*string{"test": aws.String("1234")}}

	err := setFieldWithType("expected", &any, "Field", awsstr)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := any.Field, "expected"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType(5, &any, "IntField", awsint)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := any.IntField, 5; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType("5", &any, "IntField", awsint)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := any.IntField, 5; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType(nil, &any, "IntField", awsint)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := any.IntField, 5; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType("first", &any, "StringArrayField", awsstringslice)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(any.StringArrayField), 1; got != want {
		t.Fatalf("len: got %d, want %d", got, want)
	}
	if got, want := aws.StringValue(any.StringArrayField[0]), "first"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType([]string{"one", "two", "three"}, &any, "StringArrayField", awsstringslice)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(any.StringArrayField), 3; got != want {
		t.Fatalf("len: got %d, want %d", got, want)
	}
	if got, want := aws.StringValue(any.StringArrayField[0]), "one"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := aws.StringValue(any.StringArrayField[1]), "two"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := aws.StringValue(any.StringArrayField[2]), "three"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType(int64(321), &any, "Int64ArrayField", awsint64slice)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(any.Int64ArrayField), 1; got != want {
		t.Fatalf("len: got %d, want %d", got, want)
	}
	if got, want := aws.Int64Value(any.Int64ArrayField[0]), int64(321); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType(567, &any, "Int64ArrayField", awsint64slice)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(any.Int64ArrayField), 1; got != want {
		t.Fatalf("len: got %d, want %d", got, want)
	}
	if got, want := aws.Int64Value(any.Int64ArrayField[0]), int64(567); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType("any", nil, "IntField", awsint)
	if err != nil {
		t.Fatal(err)
	}

	err = setFieldWithType("true", &any, "BooleanValueField", awsboolattribute)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := aws.BoolValue(any.BooleanValueField.Value), true; got != want {
		t.Fatalf("len: got %t, want %t", got, want)
	}
	err = setFieldWithType(nil, &any, "BooleanValueField", awsboolattribute)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := aws.BoolValue(any.BooleanValueField.Value), true; got != want {
		t.Fatalf("len: got %t, want %t", got, want)
	}
	err = setFieldWithType(false, &any, "BooleanValueField", awsboolattribute)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := aws.BoolValue(any.BooleanValueField.Value), false; got != want {
		t.Fatalf("len: got %t, want %t", got, want)
	}

	err = setFieldWithType("abcd", &any, "StringValueField", awsstringattribute)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := aws.StringValue(any.StringValueField.Value), "abcd"; got != want {
		t.Fatalf("len: got %s, want %s", got, want)
	}
	err = setFieldWithType(nil, &any, "StringValueField", awsstringattribute)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := aws.StringValue(any.StringValueField.Value), "abcd"; got != want {
		t.Fatalf("len: got %s, want %s", got, want)
	}

	err = setFieldWithType(true, &any, "BoolField", awsbool)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := any.BoolField, true; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	err = setFieldWithType(false, &any, "BoolField", awsbool)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := any.BoolField, false; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType("true", &any, "BoolPointerField", awsbool)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := *any.BoolPointerField, true; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	err = setFieldWithType(false, &any, "BoolPointerField", awsbool)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := *any.BoolPointerField, false; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType("fieldValue", &any, "StructAttribute.Str", awsstr)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := *any.StructAttribute.Str, "fieldValue"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	err = setFieldWithType("abc", &any, "MapAttribute[Field1]", awsstringpointermap)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := *any.MapAttribute["Field1"], "abc"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	err = setFieldWithType("def", &any, "MapAttribute[Field2]", awsstringpointermap)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := *any.MapAttribute["Field1"], "abc"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := *any.MapAttribute["Field2"], "def"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	err = setFieldWithType("abcd", &any, "EmptyMapAttribute[Field1]", awsstringpointermap)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := *any.EmptyMapAttribute["Field1"], "abcd"; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	err = setFieldWithType("tata", &any, "SliceStructPointerAttribute[0]Str1", awsslicestruct)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := *any.SliceStructPointerAttribute[0].Str1, "tata"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	err = setFieldWithType("toto", &any, "SliceStructPointerAttribute[0]Str2", awsslicestruct)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := *any.SliceStructPointerAttribute[0].Str2, "toto"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
