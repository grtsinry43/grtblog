"use client"

import React, {useEffect, useState, useRef} from "react"
import Image from "next/image"
import ZoomedImage from "./ZoomedImage"
import useImageZoom from "@/hooks/useImageZoom"

export default function ImageView({src, alt}: { src: string; alt: string }) {
    const [glowColor, setGlowColor] = useState("rgba(var(--foreground), 0.3)")
    const {isOpen, triggerElement, openZoom, closeZoom} = useImageZoom()
    const imgRef = useRef<HTMLImageElement>(null)
    const [isMounted, setIsMounted] = useState(false)

    useEffect(() => {
        setIsMounted(true)
    }, [])

    useEffect(() => {
        const img = imgRef.current
        if (!img || !isMounted) return

        const handleLoad = () => {
            const canvas = document.createElement("canvas")
            const ctx = canvas.getContext("2d")

            if (!ctx) return

            canvas.width = img.naturalWidth
            canvas.height = img.naturalHeight
            ctx.drawImage(img, 0, 0, img.naturalWidth, img.naturalHeight)

            try {
                const edgeColors = [
                    ctx.getImageData(0, 0, 1, 1).data,
                    ctx.getImageData(img.naturalWidth - 1, 0, 1, 1).data,
                    ctx.getImageData(0, img.naturalHeight - 1, 1, 1).data,
                    ctx.getImageData(img.naturalWidth - 1, img.naturalHeight - 1, 1, 1).data,
                ]

                const averageColor = edgeColors
                    .reduce(
                        (acc, color) => {
                            return acc.map((v, i) => v + color[i])
                        },
                        [0, 0, 0, 0],
                    )
                    .map((v) => Math.round(v / edgeColors.length))

                setGlowColor(`rgba(${averageColor[0]}, ${averageColor[1]}, ${averageColor[2]}, 0.5)`)
            } catch (e) {
                console.error("CORS error or image data not accessible", e)
            }
        }

        img.addEventListener("load", handleLoad)
        if (img.complete) handleLoad()
        return () => img.removeEventListener("load", handleLoad)
    }, [src, isMounted])

    const handleImageClick = () => {
        if (imgRef.current) {
            openZoom(imgRef.current)
        }
    }

    return (
        <>
            <Image
                ref={imgRef}
                src={src || "/placeholder.svg"}
                alt={alt}
                width={700}
                height={400}
                style={{
                    margin: "2em auto",
                    objectFit: "cover",
                    borderRadius: "5px",
                    filter: `drop-shadow(0 0 10px ${glowColor})`,
                    transition: "filter 0.3s ease-in-out, transform 0.3s ease-in-out",
                    cursor: "pointer",
                    transform: "scale(1)",
                }}
                priority
                crossOrigin="anonymous"
                onClick={handleImageClick}
                onMouseEnter={(e) => (e.currentTarget.style.transform = "scale(1.01)")}
                onMouseLeave={(e) => (e.currentTarget.style.transform = "scale(1)")}
            />
            {isOpen && (
                <ZoomedImage 
                    src={src} 
                    alt={alt} 
                    onClose={closeZoom}
                    triggerElement={triggerElement}
                />
            )}
        </>
    )
}

