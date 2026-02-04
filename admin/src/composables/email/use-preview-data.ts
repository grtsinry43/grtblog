import type { AdminEventFieldResp } from '@/services/events'

export function usePreviewData() {
    function generatePreviewData(fields: AdminEventFieldResp[]) {
        const data: Record<string, any> = {
            Name: 'Preview User',
            SiteURL: 'https://example.com'
        }

        fields.forEach((field) => {
            // Generate dummy data based on type or name
            let value: any = `[${field.name}]`

            // Simple heuristic for types
            const lowerName = field.name.toLowerCase()
            const lowerType = (field.type || '').toLowerCase()

            if (lowerType === 'int' || lowerType === 'integer' || lowerType === 'number') {
                value = 123
                if (lowerName.includes('id')) value = 1
                if (lowerName.includes('count')) value = 10
            } else if (lowerType === 'bool' || lowerType === 'boolean') {
                value = true
            } else if (lowerType === 'array' || lowerType.includes('[]')) {
                value = []
            } else if (lowerType === 'object' || lowerType === 'map') {
                value = {}
            } else {
                // String heuristics
                if (lowerName.includes('url')) value = 'https://example.com/test'
                if (lowerName.includes('email')) value = 'test@example.com'
                if (lowerName.includes('name')) value = 'Test Name'
                if (lowerName.includes('image') || lowerName.includes('avatar')) value = 'https://via.placeholder.com/150'
                if (lowerName.includes('title')) value = 'Test Generic Title'
                if (lowerName.includes('description')) value = 'This is a test description for preview.'
            }

            data[field.name] = value
        })

        return JSON.stringify(data, null, 2)
    }

    return {
        generatePreviewData
    }
}
