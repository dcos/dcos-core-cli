/*
 * DC/OS
 *
 * DC/OS API
 *
 * API version: 1.0.0
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package dcos

type CosmosPackageDescribeV1Request struct {
	PackageName    string `json:"packageName"`
	PackageVersion string `json:"packageVersion,omitempty"`
}
