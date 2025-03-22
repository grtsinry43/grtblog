"use client";

import React, {useState} from 'react';
import {AnimatePresence, motion} from "framer-motion";
import {createPortal} from "react-dom";
import {
    ArrowLeftIcon,
    ArrowRightIcon,
    CalendarIcon,
} from "@radix-ui/react-icons";
import {Camera, MapPinIcon} from "lucide-react";
import {PhotoPreview} from "@/components/album/AlbumFlowClient";

const AlbumView = ({photo, isOpen, setIsOpen, prevHandler, nextHandler, prevDisabled, nextDisabled}: {
    photo: PhotoPreview
    isOpen: boolean,
    setIsOpen: (isOpen: boolean) => void,
    prevHandler: () => void,
    nextHandler: () => void,
    prevDisabled: boolean,
    nextDisabled: boolean
}) => {
    const [warning, setWarning] = useState("");
    const [imageLoaded, setImageLoaded] = useState(false);
    return createPortal(
        <AnimatePresence>
            {isOpen && (
                <motion.div
                    initial={{opacity: 0}}
                    animate={{opacity: 1}}
                    exit={{opacity: 0}}
                    
                    
                    className="fixed inset-0 z-50 bg-black/70 backdrop-blur-xl flex items-center justify-center p-4"
                    onClick={() => setIsOpen(false)}
                >
                    <motion.div
                        initial={{scale: 0.95}}
                        animate={{scale: 1}}
                        exit={{scale: 0.95}}
                        style={{maxWidth: '80vw'}}
                        
                        
                        className="relative max-h-[90vh] bg-neutral-900 overflow-hidden flex flex-col md:flex-row lg:flex-row"
                        onClick={(e: React.MouseEvent) => e.stopPropagation()}
                    >
                        {/* 大图展示 */}
                        <div className="relative aspect-[4/3] bg-gray-800/70 backdrop-blur" style={{
                            userSelect: 'none',
                        }}>
                            <div
                                className="absolute inset-0 bg-center bg-cover"
                                style={{
                                    background: imageLoaded ? `url(${photo.url})` : photo.shade,
                                    filter: 'blur(80px)',
                                    zIndex: -1,
                                    backgroundSize: 'cover'
                                }}
                            />
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
                                
                                
                                className="w-full h-full flex justify-center"
                            >
                                {/* eslint-disable-next-line @next/next/no-img-element */}
                                <img
                                    src={photo.url}
                                    alt={`Photo from ${photo.date}`}
                                    className="w-full h-full object-contain"
                                    onLoad={
                                        () => {
                                            setImageLoaded(true)
                                            console.log("Image loaded")
                                        }
                                    }
                                />
                            </motion.div>
                            <ArrowLeftIcon
                                className="absolute top-[50%] left-4 text-white w-8 h-8 cursor-pointer opacity-35 z-10"
                                style={{transform: 'translateY(-50%)'}}
                                onClick={() => {
                                    if (!prevDisabled) {
                                        prevHandler()
                                    } else {
                                        setWarning('已经是本组第一张了捏')
                                        setTimeout(() => {
                                            setWarning('')
                                        }, 500)
                                    }
                                }}

                            />
                            <ArrowRightIcon
                                className="absolute top-[50%] right-4 text-white w-8 h-8 cursor-pointer opacity-35 z-10"
                                style={{transform: 'translateY(-50%)'}}
                                onClick={() => {
                                    if (!nextDisabled) {
                                        nextHandler()
                                    } else {
                                        setWarning('已经是本组最后一张了捏')
                                        setTimeout(() => {
                                            setWarning('')
                                        }, 500)
                                    }
                                }}
                            />
                            {warning &&
                                <div
                                    className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-10">
                                    <div className="p-4 bg-black/40 backdrop-blur-3xl rounded-lg text-white">
                                        {warning}
                                    </div>
                                </div>
                            }
                        </div>

                        {/*    <img*/}
                        {/*        src={photo.url}*/}
                        {/*        alt={`Photo from ${photo.date}`}*/}
                        {/*        className="w-full h-full object-contain"*/}
                        {/*        style={{position: 'relative'}}*/}
                        {/*    />*/}

                        {/* 信息面板 */}
                        <div className="p-6 space-y-4 text-sm text-gray-300 flex flex-col min-w-[300px]">
                            <div className="flex justify-between items-start">
                                <h3 className="text-lg font-semibold text-white"> 信息 </h3>
                                <button
                                    onClick={() => setIsOpen(false)}
                                    className="p-2 hover:text-white transition-colors"
                                >
                                    ✕
                                </button>
                            </div>

                            <div className="flex flex-col gap-4 text-sm">
                                <div className="flex items-center space-x-2">
                                    <CalendarIcon className="w-4 h-4 text-gray-400"/>
                                    <span className="text-gray-400"> 拍摄时间:</span>
                                    <span>{new Date(photo.date).toLocaleString()}</span>
                                </div>
                                {photo.location && (
                                    <div className="flex items-center space-x-2">
                                        <MapPinIcon className="w-4 h-4 text-gray-400"/>
                                        <span className="text-gray-400"> 拍摄地点:</span>
                                        <span>{photo.location}</span>
                                    </div>
                                )}
                                {photo.device && (
                                    <div className="flex items-center space-x-2">
                                        <Camera className="w-4 h-4 text-gray-400"/>
                                        <span className="text-gray-400"> 拍摄设备:</span>
                                        <span>{photo.device}</span>
                                    </div>
                                )}
                            </div>

                            {photo.description && (
                                <div className="prose prose-invert max-w-none pt-12">
                                    {photo.description}
                                </div>
                            )}

                        </div>
                    </motion.div>
                </motion.div>
            )}
        </AnimatePresence>,
        document.body
    );
};

export default AlbumView;
