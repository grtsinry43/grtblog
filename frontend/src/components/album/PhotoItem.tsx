"use client"

import {forwardRef, useState} from "react"
import {motion, AnimatePresence} from "framer-motion"
import type {PhotoPreview} from "@/components/album/AlbumFlowClient"

interface PhotoItemProps {
    photo: PhotoPreview,
    setIsOpened: (isOpen: boolean) => void
}

const PhotoItem = forwardRef<HTMLDivElement, PhotoItemProps>(({photo, setIsOpened}, ref) => {
    const [imageLoaded, setImageLoaded] = useState(false)

    return (
        <motion.div
            ref={ref}
            initial={{opacity: 0, scale: 0.9}}
            animate={{opacity: 1, scale: 1}}
            transition={{type: "spring", stiffness: 300, damping: 20}}
            
            
            className="relative aspect-[4/3] overflow-hidden rounded-lg shadow-md hover:shadow-lg transition-shadow duration-300"
            style={{backgroundColor: photo.shade}}
            onClick={() => setIsOpened(true)}
        >
            <AnimatePresence>
                {!imageLoaded && (
                    <motion.div
                        initial={{opacity: 1}}
                        exit={{opacity: 0}}
                        transition={{duration: 0.5}}
                        
                        
                        className="absolute inset-0 flex items-center justify-center"
                    >
                        <div
                            className="w-10 h-10 border-4 border-white border-t-transparent rounded-full animate-spin"></div>
                    </motion.div>
                )}
            </AnimatePresence>
            <motion.div
                initial={{
                    clipPath: "polygon(0 0, 0 0, 0 0)",
                    filter: "blur(20px)",
                }}
                animate={
                    imageLoaded
                        ? {
                            clipPath: "polygon(0 0, 100% 0, 100% 100%, 0 100%)",
                            filter: "blur(0px)",
                        }
                        : {}
                }
                transition={{
                    clipPath: {duration: 1, ease: "easeInOut"},
                    filter: {duration: 0.8, ease: "easeOut"},
                }}
                
                
                className="w-full h-full"
            >
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                    src={photo.url}
                    alt={`Photo from ${photo.date}`}
                    className="w-full h-full object-cover transition-transform duration-300 hover:scale-105"
                    onLoad={
                        () => {
                            setImageLoaded(true)
                            console.log("Image loaded")
                        }
                    }
                />
            </motion.div>
        </motion.div>

    )
})

PhotoItem.displayName = "PhotoItem"

export default PhotoItem

