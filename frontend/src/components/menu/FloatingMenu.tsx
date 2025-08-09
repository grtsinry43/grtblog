'use client'

import * as React from 'react'
import {motion, AnimatePresence} from 'framer-motion'
import {Settings, Palette, Droplet} from 'lucide-react'

import {Button} from '@/components/ui/button'
import {Switch} from '@/components/ui/switch'
import {Label} from '@/components/ui/label'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import {ThemeSelector} from './ThemeSelector'
import {ColorPicker} from "@/components/menu/ColorPicker";
import {useEffect} from "react";

export interface MenuItem {
    icon: React.ReactNode
    label: string
    type: 'switch' | 'select' | 'button'
    value?: string | boolean
    onClick?: (e?: boolean) => void
}

export default function FloatingMenu({items}: { items: MenuItem[] }) {
    const [isOpen, setIsOpen] = React.useState(false)

    useEffect(() => {
        if (localStorage.getItem('primary-color')) document.documentElement.style.setProperty('--primary', localStorage.getItem('primary-color'))
    }, [isOpen])

    return (
        <div className="fixed bottom-4 right-4 z-50">
            <AnimatePresence>
                {isOpen && (
                    <motion.div
                        initial={{opacity: 0, scale: 0.8, y: 20}}
                        animate={{opacity: 1, scale: 1, y: 0}}
                        exit={{opacity: 0, scale: 0.8, y: 20}}
                        transition={{type: 'spring', stiffness: 300, damping: 30}}
                        
                        
                        className="absolute bottom-16 right-0 w-72 rounded-lg border bg-background shadow-lg overflow-hidden"
                    >
                        <div className="absolute inset-0 bg-background/80 backdrop-blur-sm z-0"/>
                        <div className="relative z-10 p-4 space-y-4">
                            {items.map((item, index) => (
                                <div key={index} className="space-y-2">
                                    <div className="flex items-center space-x-2">
                                        {item.icon}
                                        <Label htmlFor={item.label.toLowerCase().replace(' ', '-')}>{item.label}</Label>
                                    </div>
                                    {item.type === 'switch' &&
                                        <Switch id={item.label.toLowerCase().replace(' ', '-')}
                                                checked={item.value as boolean}
                                                onCheckedChange={item.onClick}
                                        />
                                    }
                                    {item.type === 'select' && (
                                        <Select>
                                            <SelectTrigger id={item.label.toLowerCase().replace(' ', '-')}>
                                                <SelectValue placeholder="Select size"/>
                                            </SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="small">Small</SelectItem>
                                                <SelectItem value="medium">Medium</SelectItem>
                                                <SelectItem value="large">Large</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    )}
                                    {item.type === 'button' &&
                                        <Button onClick={() => item.onClick}>{item.label}</Button>}
                                </div>
                            ))}
                            <div className="space-y-2">
                                <div className="flex items-center space-x-2">
                                    <Palette className="h-4 w-4"/>
                                    <Label> 设置主题 </Label>
                                </div>
                                <ThemeSelector/>
                            </div>
                            <div className="space-y-2">
                                <div className="flex items-center space-x-2">
                                    <Droplet className="h-4 w-4"/>
                                    <Label> 主色调 </Label>
                                </div>
                                <ColorPicker/>
                            </div>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
            <Button
                size="icon"
                className="rounded-full bg-background text-foreground shadow-sm"
                style={{
                    border: '1px solid rgba(var(--foreground), 0.1)',
                }}
                onClick={() => setIsOpen(!isOpen)}
            >
                <Settings className="h-6 w-6"/>
            </Button>
        </div>
    )
}
