/*
 * DC/OS
 *
 * DC/OS API
 *
 * API version: 1.0.0
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package dcos

// Defines the behavior if a task fails
type MetronomeV1JobRunRestart struct {
	// The policy to use if a job fails. NEVER will never try to relaunch a job. ON_FAILURE will try to start a job in case of failure.
	Policy string `json:"policy"`
	// If the job fails, how long should we try to restart the job. If no value is set, this means forever.
	ActiveDeadlineSeconds int32 `json:"activeDeadlineSeconds,omitempty"`
}
