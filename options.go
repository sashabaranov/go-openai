// Copyright (c) 2012-2023 Grabtaxi Holdings PTE LTD (GRAB), All Rights Reserved. NOTICE: All information contained herein
// is, and remains the property of GRAB. The intellectual and technical concepts contained herein are confidential, proprietary
// and controlled by GRAB and may be covered by patents, patents in process, and are protected by trade secret or copyright law.
//
// You are strictly forbidden to copy, download, store (in any medium), transmit, disseminate, adapt or change this material
// in any way unless prior written permission is obtained from GRAB. Access to the source code contained herein is hereby
// forbidden to anyone except current GRAB employees or contractors with binding Confidentiality and Non-disclosure agreements
// explicitly covering such access.
//
// The copyright notice above does not evidence any actual or intended publication or disclosure of this source code,
// which includes information that is confidential and/or proprietary, and is a trade secret, of GRAB.
//
// ANY REPRODUCTION, MODIFICATION, DISTRIBUTION, PUBLIC PERFORMANCE, OR PUBLIC DISPLAY OF OR THROUGH USE OF THIS SOURCE
// CODE WITHOUT THE EXPRESS WRITTEN CONSENT OF GRAB IS STRICTLY PROHIBITED, AND IN VIOLATION OF APPLICABLE LAWS AND
// INTERNATIONAL TREATIES. THE RECEIPT OR POSSESSION OF THIS SOURCE CODE AND/OR RELATED INFORMATION DOES NOT CONVEY
// OR IMPLY ANY RIGHTS TO REPRODUCE, DISCLOSE OR DISTRIBUTE ITS CONTENTS, OR TO MANUFACTURE, USE, OR SELL ANYTHING
// THAT IT MAY DESCRIBE, IN WHOLE OR IN PART.

package openai

type ConfigOption func(config *ClientConfig)

// WithBaseURL configures base url which should start with "https", e.g. https://exmample.com
func WithBaseURL(baseURL string) ConfigOption {
	return func(config *ClientConfig) {
		config.BaseURL = baseURL
	}
}

func WithAPIType(apiType APIType) ConfigOption {
	return func(config *ClientConfig) {
		config.APIType = apiType
	}
}

func WithAPIVersion(apiVersion string) ConfigOption {
	return func(config *ClientConfig) {
		config.APIVersion = apiVersion
	}
}

func WithOrgID(orgID string) ConfigOption {
	return func(config *ClientConfig) {
		config.OrgID = orgID
	}
}

func WithModelMapperFunc(mapper func(model string) string) ConfigOption {
	return func(config *ClientConfig) {
		config.AzureModelMapperFunc = mapper
	}
}

func WithEmptyMessagesLimit(limit uint) ConfigOption {
	return func(config *ClientConfig) {
		config.EmptyMessagesLimit = limit
	}
}
