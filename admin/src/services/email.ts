import { request } from './http'


export interface EmailTemplate {
    id: number
    code: string
    name: string
    eventName: string
    subjectTemplate: string
    htmlTemplate: string
    textTemplate: string
    toEmails: string[]
    isEnabled: boolean
    createdAt: string
    updatedAt: string
}

export interface CreateEmailTemplateReq {
    code: string
    name: string
    eventName: string
    subjectTemplate: string
    htmlTemplate: string
    textTemplate: string
    toEmails: string[]
    isEnabled: boolean
}

export interface UpdateEmailTemplateReq {
    name: string
    eventName: string
    subjectTemplate: string
    htmlTemplate: string
    textTemplate: string
    toEmails: string[]
    isEnabled: boolean
}

export interface EmailTemplatePreviewReq {
    variables: Record<string, any>
}

export interface EmailTemplatePreviewResp {
    subject: string
    htmlBody: string
    textBody: string
}

export interface EmailTemplateTestReq {
    toEmails: string[]
    variables: Record<string, any>
}

export interface EmailSubscription {
    id: number
    email: string
    eventName: string
    status: 'pending' | 'verified' | 'unsubscribed'
    sourceIp: string
    unsubscribedAt?: string
    token?: string
    createdAt: string
    updatedAt: string
}

export interface EmailSubscriptionListResp {
    items: EmailSubscription[]
    total: number
    page: number
    size: number
}

export interface EmailSubscriptionListParams {
    page?: number
    pageSize?: number
    eventName?: string
    status?: string
    search?: string
}

export interface BatchUpdateEmailSubscriptionStatusReq {
    ids: number[]
    status: string
}

function stripEmpty<T extends Record<string, any>>(value: T) {
    return Object.fromEntries(
        Object.entries(value).filter(([, entry]) => entry !== undefined && entry !== null && entry !== ''),
    ) as T
}

// 邮件事件相关 API

// 邮件模板相关 API
export function listEmailTemplates() {
    return request<EmailTemplate[]>('/admin/email/templates', {
        method: 'GET',
    })
}

export function createEmailTemplate(data: CreateEmailTemplateReq) {
    return request<EmailTemplate>('/admin/email/templates', {
        method: 'POST',
        body: data,
    })
}

export function updateEmailTemplate(code: string, data: UpdateEmailTemplateReq) {
    return request<EmailTemplate>(`/admin/email/templates/${code}`, {
        method: 'PUT',
        body: data,
    })
}

export function deleteEmailTemplate(code: string) {
    return request<void>(`/admin/email/templates/${code}`, {
        method: 'DELETE',
    })
}

export function previewEmailTemplate(code: string, data: EmailTemplatePreviewReq) {
    return request<EmailTemplatePreviewResp>(`/admin/email/templates/${code}/preview`, {
        method: 'POST',
        body: data,
    })
}

export function testEmailTemplate(code: string, data: EmailTemplateTestReq) {
    return request<void>(`/admin/email/templates/${code}/test`, {
        method: 'POST',
        body: data,
    })
}

export function getTemplateByCode(code: string) {
    return request<EmailTemplate>(`/admin/email/templates/${code}`, {
        method: 'GET',
    })
}

// 邮件订阅相关 API
export function listEmailSubscriptions(params: EmailSubscriptionListParams) {
    return request<EmailSubscriptionListResp>('/admin/email/subscriptions', {
        method: 'GET',
        query: stripEmpty(params),
    })
}

export function batchUpdateEmailSubscriptionStatus(data: BatchUpdateEmailSubscriptionStatusReq) {
    return request<void>('/admin/email/subscriptions/status', {
        method: 'PUT',
        body: data,
    })
}
