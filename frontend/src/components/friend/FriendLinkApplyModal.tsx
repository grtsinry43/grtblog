"use client";

import type React from "react"
import {useEffect, useState} from "react"
import {AnimatePresence, motion, useDragControls} from "framer-motion"
import {X} from "lucide-react"
import {clsx} from "clsx"
import {varela_round, noto_sans_sc} from "@/app/fonts/font"
import {useAppSelector} from "@/redux/hooks"
import {applyFriendLink} from "@/api/friends"
import {Button} from "@/components/ui/button";
import {TextField} from "@radix-ui/themes"
import {createPortal} from "react-dom";

const FriendLinkApplyModal = ({isOpen, onClose}: { isOpen: boolean; onClose: () => void }) => {
    const [isClient, setIsClient] = useState(false);
    const [form, setForm] = useState({
        siteName: "",
        siteUrl: "",
        siteDescription: "",
        avatarUrl: "",
    })
    const [error, setError] = useState("")
    const dragControls = useDragControls()

    const submitForm = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        if (!form.siteName || !form.siteUrl || !form.siteDescription || !form.avatarUrl) {
            setError("请填写所有必填字段")
            return
        }
        applyFriendLink({
            name: form.siteName,
            url: form.siteUrl,
            logo: form.avatarUrl,
            description: form.siteDescription,
        }).then((res) => {
            if (!res) {
                setError("申请失败，请稍后重试")
            } else {
                onClose()
            }
        })
    }

    const isLogin = useAppSelector((state) => state.user.isLogin)

    useEffect(() => {
        if (!isLogin) {
            setError("请先登录后再申请友链哦")
        } else {
            setError("")
        }
    }, [isLogin])

    useEffect(() => {
        setIsClient(true);
    }, []);

    if (!isClient) {
        return null;
    }

    return createPortal(
        <AnimatePresence>
            {isOpen && (
                <div className="fixed inset-0 z-50 flex items-center justify-center">
                    <motion.div
                        initial={{opacity: 0}}
                        animate={{opacity: 1}}
                        exit={{opacity: 0}}
                        
                        
                        className="absolute inset-0 bg-black/30 backdrop-blur-sm"
                        onClick={onClose}
                    />
                    <motion.div
                        drag
                        dragControls={dragControls}
                        dragMomentum={false}
                        initial={{scale: 0.9, opacity: 0}}
                        animate={{scale: 1, opacity: 1}}
                        exit={{scale: 0.9, opacity: 0}}
                        
                        
                        className="relative w-full max-w-md rounded-lg bg-white/80 p-6 shadow-xl dark:bg-gray-800/80"
                    >
                        <div
                            className="absolute inset-x-0 top-0 h-14 cursor-move rounded-t-lg bg-gray-100 dark:bg-gray-700"
                            onPointerDown={(e) => dragControls.start(e)}
                        >
                            <h2
                                className={clsx(
                                    "pt-4 text-center text-xl font-semibold",
                                    varela_round.className,
                                    noto_sans_sc.className,
                                )}
                            >
                                申请友链
                            </h2>
                        </div>
                        <Button variant="ghost" size="icon" className="absolute right-2 top-2" onClick={onClose}>
                            <X className="h-4 w-4"/>
                        </Button>
                        <div className="mt-14">
                            <p className="mb-4 text-center text-sm font-medium text-gray-600 dark:text-gray-300">
                                愿岁并谢，与长友兮
                            </p>
                            <form onSubmit={submitForm} className="space-y-4">
                                {error && (
                                    <div
                                        className="rounded-md bg-red-50 p-3 text-sm text-red-700 dark:bg-red-900/50 dark:text-red-200">
                                        {error}
                                    </div>
                                )}
                                <div className="space-y-2">
                                    <label className="text-sm font-medium"> 网站名称 </label>
                                    <TextField.Root
                                        value={form.siteName}
                                        style={{
                                            backgroundColor: 'rgba(var(--foreground), 0.03)',
                                            flex: '1',
                                            boxShadow: 'none',
                                            minHeight: '2rem',
                                            border: '1px solid rgba(var(--foreground), 0.1)',
                                            borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                            outline: 'none',
                                            borderRadius: '0.375rem',
                                            padding: '0 0.5rem',
                                        }}
                                        onChange={(e) => setForm({...form, siteName: e.target.value})}
                                        placeholder="输入您的网站名称"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium"> 网站 URL</label>
                                    <TextField.Root
                                        value={form.siteUrl}
                                        style={{
                                            backgroundColor: 'rgba(var(--foreground), 0.03)',
                                            flex: '1',
                                            boxShadow: 'none',
                                            minHeight: '2rem',
                                            border: '1px solid rgba(var(--foreground), 0.1)',
                                            borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                            outline: 'none',
                                            borderRadius: '0.375rem',
                                            padding: '0 0.5rem',
                                        }}
                                        onChange={(e) => setForm({...form, siteUrl: e.target.value})}
                                        placeholder="https://example.com"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium"> 网站描述 </label>
                                    <TextField.Root
                                        value={form.siteDescription}
                                        style={{
                                            backgroundColor: 'rgba(var(--foreground), 0.03)',
                                            flex: '1',
                                            boxShadow: 'none',
                                            minHeight: '2rem',
                                            border: '1px solid rgba(var(--foreground), 0.1)',
                                            borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                            outline: 'none',
                                            borderRadius: '0.375rem',
                                            padding: '0 0.5rem',
                                        }}
                                        onChange={(e) => setForm({...form, siteDescription: e.target.value})}
                                        placeholder="简短描述您的网站"
                                    />
                                </div>
                                <div className="space-y-2">
                                    <label className="text-sm font-medium"> 头像 URL</label>
                                    <TextField.Root
                                        value={form.avatarUrl}
                                        style={{
                                            backgroundColor: 'rgba(var(--foreground), 0.03)',
                                            flex: '1',
                                            boxShadow: 'none',
                                            minHeight: '2rem',
                                            border: '1px solid rgba(var(--foreground), 0.1)',
                                            borderBottom: '1px solid rgba(var(--foreground), 0.5)',
                                            outline: 'none',
                                            borderRadius: '0.375rem',
                                            padding: '0 0.5rem',
                                        }}
                                        onChange={(e) => setForm({...form, avatarUrl: e.target.value})}
                                        placeholder="https://example.com/avatar.png"
                                    />
                                </div>
                                <div className="flex justify-end space-x-2 pt-4">
                                    <Button variant="outline" onClick={onClose}>
                                        取消
                                    </Button>
                                    <Button type="submit"> 提交申请 </Button>
                                </div>
                            </form>
                        </div>
                    </motion.div>
                </div>
            )}
        </AnimatePresence>,
        window.document.body,
    )
}

export default FriendLinkApplyModal

