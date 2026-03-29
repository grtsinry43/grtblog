import ExifReader from 'exifreader'
import type { PhotoExif } from '@/services/albums'

/**
 * 从图片文件中提取 EXIF 信息。
 */
export async function extractExif(file: File): Promise<PhotoExif | null> {
  try {
    const buffer = await file.arrayBuffer()
    const tags = ExifReader.load(buffer, { expanded: true })

    const exif: PhotoExif = {}

    // 设备
    if (tags.exif?.Make?.description) exif.make = tags.exif.Make.description
    if (tags.exif?.Model?.description) exif.model = tags.exif.Model.description
    if (tags.exif?.LensModel?.description) exif.lensModel = tags.exif.LensModel.description

    // 拍摄参数
    if (tags.exif?.FocalLength?.description) exif.focalLength = tags.exif.FocalLength.description
    if (tags.exif?.FNumber?.description) exif.fNumber = tags.exif.FNumber.description
    if (tags.exif?.ExposureTime?.description) exif.exposureTime = tags.exif.ExposureTime.description
    if (tags.exif?.ISOSpeedRatings?.description) exif.iso = Number(tags.exif.ISOSpeedRatings.description)

    // GPS
    if (tags.gps?.Latitude != null) exif.gpsLatitude = tags.gps.Latitude
    if (tags.gps?.Longitude != null) exif.gpsLongitude = tags.gps.Longitude
    if (tags.gps?.Altitude != null) exif.gpsAltitude = tags.gps.Altitude

    // 时间
    if (tags.exif?.DateTimeOriginal?.description) exif.dateTimeOriginal = tags.exif.DateTimeOriginal.description

    // 尺寸
    const w = tags.file?.['Image Width']?.value ?? tags.exif?.PixelXDimension?.value
    const h = tags.file?.['Image Height']?.value ?? tags.exif?.PixelYDimension?.value
    if (w) exif.imageWidth = Number(w)
    if (h) exif.imageHeight = Number(h)

    // 方向
    const orientation = (tags.file as Record<string, any>)?.Orientation?.value
    if (orientation != null) exif.orientation = Number(orientation)

    return Object.keys(exif).length > 0 ? exif : null
  } catch {
    return null
  }
}

/**
 * 从 EXIF 中提取设备描述。
 */
export function exifDevice(exif?: PhotoExif | null): string | undefined {
  if (!exif) return undefined
  const parts = [exif.make, exif.model].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : undefined
}

/**
 * 从 EXIF 中提取拍摄参数摘要。
 */
export function exifShootingInfo(exif?: PhotoExif | null): string | undefined {
  if (!exif) return undefined
  const parts: string[] = []
  if (exif.focalLength) parts.push(exif.focalLength)
  if (exif.fNumber) parts.push(`f/${exif.fNumber}`)
  if (exif.exposureTime) parts.push(`${exif.exposureTime}s`)
  if (exif.iso) parts.push(`ISO ${exif.iso}`)
  return parts.length > 0 ? parts.join('  ') : undefined
}

/**
 * 从 EXIF GPS 坐标生成简要位置文本。
 */
export function exifLocation(exif?: PhotoExif | null): string | undefined {
  if (!exif?.gpsLatitude || !exif?.gpsLongitude) return undefined
  return `${exif.gpsLatitude.toFixed(4)}, ${exif.gpsLongitude.toFixed(4)}`
}

