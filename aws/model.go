/*
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

import "github.com/wallix/awless/graph"

var awsResourcesDef = map[graph.ResourceType]map[string]*propertyTransform{
	//EC2
	graph.Instance: {
		"Id":             {name: "InstanceId", transform: extractValueFn},
		"Name":           {name: "Tags", transform: extractTagFn("Name")},
		"Type":           {name: "InstanceType", transform: extractValueFn},
		"SubnetId":       {name: "SubnetId", transform: extractValueFn},
		"VpcId":          {name: "VpcId", transform: extractValueFn},
		"PublicIp":       {name: "PublicIpAddress", transform: extractValueFn},
		"PrivateIp":      {name: "PrivateIpAddress", transform: extractValueFn},
		"ImageId":        {name: "ImageId", transform: extractValueFn},
		"LaunchTime":     {name: "LaunchTime", transform: extractValueFn},
		"State":          {name: "State", transform: extractFieldFn("Name")},
		"KeyName":        {name: "KeyName", transform: extractValueFn},
		"SecurityGroups": {name: "SecurityGroups", transform: extractSliceValues("GroupId")},
	},
	graph.Vpc: {
		"Id":        {name: "VpcId", transform: extractValueFn},
		"Name":      {name: "Tags", transform: extractTagFn("Name")},
		"IsDefault": {name: "IsDefault", transform: extractValueFn},
		"State":     {name: "State", transform: extractValueFn},
		"CidrBlock": {name: "CidrBlock", transform: extractValueFn},
	},
	graph.Subnet: {
		"Id":                  {name: "SubnetId", transform: extractValueFn},
		"Name":                {name: "Tags", transform: extractTagFn("Name")},
		"VpcId":               {name: "VpcId", transform: extractValueFn},
		"MapPublicIpOnLaunch": {name: "MapPublicIpOnLaunch", transform: extractValueFn},
		"State":               {name: "State", transform: extractValueFn},
		"CidrBlock":           {name: "CidrBlock", transform: extractValueFn},
		"AvailabilityZone":    {name: "AvailabilityZone", transform: extractValueFn},
		"DefaultForAz":        {name: "DefaultForAz", transform: extractValueFn},
	},
	graph.SecurityGroup: {
		"Id":            {name: "GroupId", transform: extractValueFn},
		"Name":          {name: "GroupName", transform: extractValueFn},
		"Description":   {name: "Description", transform: extractValueFn},
		"InboundRules":  {name: "IpPermissions", transform: extractIpPermissionSliceFn},
		"OutboundRules": {name: "IpPermissionsEgress", transform: extractIpPermissionSliceFn},
		"OwnerId":       {name: "OwnerId", transform: extractValueFn},
		"VpcId":         {name: "VpcId", transform: extractValueFn},
	},
	graph.Keypair: {
		"Id":             {name: "KeyName", transform: extractValueFn},
		"Name":           {name: "KeyName", transform: extractValueFn},
		"KeyFingerprint": {name: "KeyFingerprint", transform: extractValueFn},
	},
	graph.Volume: {
		"Id":               {name: "VolumeId", transform: extractValueFn},
		"Name":             {name: "Tags", transform: extractTagFn("Name")},
		"VolumeType":       {name: "VolumeType", transform: extractValueFn},
		"State":            {name: "State", transform: extractValueFn},
		"Size":             {name: "Size", transform: extractValueFn},
		"Encrypted":        {name: "Encrypted", transform: extractValueFn},
		"CreateTime":       {name: "CreateTime", transform: extractTimeFn},
		"AvailabilityZone": {name: "AvailabilityZone", transform: extractValueFn},
	},
	graph.InternetGateway: {
		"Id":   {name: "InternetGatewayId", transform: extractValueFn},
		"Name": {name: "Tags", transform: extractTagFn("Name")},
		"Vpcs": {name: "Attachments", transform: extractSliceValues("VpcId")},
	},
	graph.RouteTable: {
		"Id":     {name: "RouteTableId", transform: extractValueFn},
		"Name":   {name: "Tags", transform: extractTagFn("Name")},
		"VpcId":  {name: "VpcId", transform: extractValueFn},
		"Routes": {name: "Routes", transform: extractRoutesSliceFn},
		"Main":   {name: "Associations", transform: extractHasATrueBoolInStructSliceFn("Main")},
	},
	graph.AvailabilityZone: {
		"Id":       {name: "ZoneName", transform: extractValueFn},
		"Name":     {name: "ZoneName", transform: extractValueFn},
		"State":    {name: "State", transform: extractValueFn},
		"Region":   {name: "RegionName", transform: extractValueFn},
		"Messages": {name: "Messages", transform: extractSliceValues("Message")},
	},
	// LoadBalancer
	graph.LoadBalancer: {
		"Id":                    {name: "LoadBalancerArn", transform: extractValueFn},
		"Name":                  {name: "LoadBalancerName", transform: extractValueFn},
		"AvailabilityZones":     {name: "AvailabilityZones", transform: extractSliceValues("ZoneName")},
		"Subnets":               {name: "AvailabilityZones", transform: extractSliceValues("SubnetId")},
		"CanonicalHostedZoneId": {name: "CanonicalHostedZoneId", transform: extractValueFn},
		"CreateTime":            {name: "CreatedTime", transform: extractTimeFn},
		"DNSName":               {name: "DNSName", transform: extractValueFn},
		"IpAddressType":         {name: "IpAddressType", transform: extractValueFn},
		"Scheme":                {name: "Scheme", transform: extractValueFn},
		"State":                 {name: "State", transform: extractFieldFn("Code")},
		"Type":                  {name: "Type", transform: extractValueFn},
		"VpcId":                 {name: "VpcId", transform: extractValueFn},
	},
	graph.TargetGroup: {
		"Id":   {name: "TargetGroupArn", transform: extractValueFn},
		"Name": {name: "TargetGroupName", transform: extractValueFn},
		"HealthCheckIntervalSeconds": {name: "HealthCheckIntervalSeconds", transform: extractValueFn},
		"HealthCheckPath":            {name: "HealthCheckPath", transform: extractValueFn},
		"HealthCheckPort":            {name: "HealthCheckPort", transform: extractValueFn},
		"HealthCheckProtocol":        {name: "HealthCheckProtocol", transform: extractValueFn},
		"HealthCheckTimeoutSeconds":  {name: "HealthCheckTimeoutSeconds", transform: extractValueFn},
		"HealthyThresholdCount":      {name: "HealthyThresholdCount", transform: extractValueFn},
		"Matcher":                    {name: "Matcher", transform: extractFieldFn("HttpCode")},
		"Port":                       {name: "Port", transform: extractValueFn},
		"Protocol":                   {name: "Protocol", transform: extractValueFn},
		"UnhealthyThresholdCount":    {name: "UnhealthyThresholdCount", transform: extractValueFn},
		"VpcId":                      {name: "VpcId", transform: extractValueFn},
	},
	graph.Listener: {
		"Id":           {name: "ListenerArn", transform: extractValueFn},
		"Certificates": {name: "Certificates", transform: extractSliceValues("CertificateArn")},
		"Actions":      {name: "DefaultActions", transform: extractSliceValues("Type")},
		"LoadBalancer": {name: "LoadBalancerArn", transform: extractValueFn},
		"Port":         {name: "Port", transform: extractValueFn},
		"Protocol":     {name: "Protocol", transform: extractValueFn},
		"SslPolicy":    {name: "SslPolicy", transform: extractValueFn},
	},
	//IAM
	graph.User: {
		"Id":                   {name: "UserId", transform: extractValueFn},
		"Name":                 {name: "UserName", transform: extractValueFn},
		"Arn":                  {name: "Arn", transform: extractValueFn},
		"Path":                 {name: "Path", transform: extractValueFn},
		"CreateDate":           {name: "CreateDate", transform: extractTimeFn},
		"PasswordLastUsedDate": {name: "PasswordLastUsed", transform: extractTimeFn},
		"InlinePolicies":       {name: "UserPolicyList", transform: extractSliceValues("PolicyName")},
	},
	graph.Role: {
		"Id":             {name: "RoleId", transform: extractValueFn},
		"Name":           {name: "RoleName", transform: extractValueFn},
		"Arn":            {name: "Arn", transform: extractValueFn},
		"CreateDate":     {name: "CreateDate", transform: extractTimeFn},
		"Path":           {name: "Path", transform: extractValueFn},
		"InlinePolicies": {name: "RolePolicyList", transform: extractSliceValues("PolicyName")},
	},
	graph.Group: {
		"Id":             {name: "GroupId", transform: extractValueFn},
		"Name":           {name: "GroupName", transform: extractValueFn},
		"Arn":            {name: "Arn", transform: extractValueFn},
		"CreateDate":     {name: "CreateDate", transform: extractTimeFn},
		"Path":           {name: "Path", transform: extractValueFn},
		"InlinePolicies": {name: "GroupPolicyList", transform: extractSliceValues("PolicyName")},
	},
	graph.Policy: {
		"Id":           {name: "PolicyId", transform: extractValueFn},
		"Name":         {name: "PolicyName", transform: extractValueFn},
		"Arn":          {name: "Arn", transform: extractValueFn},
		"CreateDate":   {name: "CreateDate", transform: extractTimeFn},
		"UpdateDate":   {name: "UpdateDate", transform: extractTimeFn},
		"Description":  {name: "Description", transform: extractValueFn},
		"IsAttachable": {name: "IsAttachable", transform: extractValueFn},
		"Path":         {name: "Path", transform: extractValueFn},
	},
	//S3
	graph.Bucket: {
		"Id":         {name: "Name", transform: extractValueFn},
		"Name":       {name: "Name", transform: extractValueFn},
		"CreateDate": {name: "CreationDate", transform: extractTimeFn},
		"Grants":     {fetch: fetchAndExtractGrantsFn},
	},
	graph.Object: {
		"Id":           {name: "Key", transform: extractValueFn},
		"Key":          {name: "Key", transform: extractValueFn},
		"ModifiedDate": {name: "LastModified", transform: extractTimeFn},
		"OwnerId":      {name: "Owner", transform: extractFieldFn("ID")},
		"Size":         {name: "Size", transform: extractValueFn},
		"Class":        {name: "StorageClass", transform: extractValueFn},
	},
	//Notification
	graph.Subscription: {
		"Id":              {name: "Endpoint", transform: extractValueFn},
		"Endpoint":        {name: "Endpoint", transform: extractValueFn},
		"Owner":           {name: "Owner", transform: extractValueFn},
		"Protocol":        {name: "Protocol", transform: extractValueFn},
		"SubscriptionArn": {name: "SubscriptionArn", transform: extractValueFn},
		"TopicArn":        {name: "TopicArn", transform: extractValueFn},
	},
	graph.Topic: {
		"Id":       {name: "TopicArn", transform: extractValueFn},
		"TopicArn": {name: "TopicArn", transform: extractValueFn},
	},
	//Queue
	graph.Queue: {}, //Manually set
}
