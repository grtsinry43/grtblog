'use client';

import React, {useState, useEffect, useCallback} from 'react';
import {FloatingPanel, Search, ConfigProvider} from 'react-vant';
import 'react-vant/lib/index.css';
import '@vant/touch-emulator';
import {Button} from '@/components/ui/button';
import {TocItem} from '@/components/article/Toc';
import {MessageCircleIcon, TableOfContentsIcon} from 'lucide-react';
import useIsMobile from '@/hooks/useIsMobile';
import {motion, AnimatePresence} from 'framer-motion';
import emitter from '@/utils/eventBus';
import {HeartIcon, PinTopIcon} from "@radix-ui/react-icons";
import {likeRequest} from "@/api/like";
import {toast} from "react-toastify";

const FloatingTocMobile = ({toc, targetId, type}: { toc: TocItem[], targetId: string, type: string }) => {
    const isMobile = useIsMobile();
    const [anchors, setAnchors] = useState([0, 0, 0]);
    const [showPanel, setShowPanel] = useState(false);
    const [activeAnchor, setActiveAnchor] = useState<string | null>(null);
    const [doms, setDoms] = useState<HTMLElement[]>([]);
    const [isUtilIconShowed, setIsUtilIconShowed] = useState(false);
    const [isTopIconShowed, setIsTopIconShowed] = useState(false);

    const showPanelHandle = () => {
        setShowPanel(true);
        document.body.style.overflow = 'hidden';
    };

    const hidePanelHandle = () => {
        setShowPanel(false);
        document.body.style.overflow = '';
    };

    const themeVars = {
        '--rv-floating-panel-z-index': 1001,
        '--rv-floating-panel-background-color': 'rgba(var(--background), 1)',
        '--rv-floating-panel-header-background-color': 'var(--rv-background)',
        '--rv-floating-panel-header-padding': '8px',
        '--rv-floating-panel-thumb-background-color': 'var(--rv-gray-5)',
        '--rv-floating-panel-thumb-width': '20px',
        '--rv-floating-panel-thumb-height': '4px',
        '--rv-search-background-color': 'var(--background)',
        '--rv-search-content-background-color': 'rgba(var(--foreground), 0.05)',
        '--rv-search-left-icon-color': 'var(--foreground)',
        '--rv-search-label-color': 'var(--foreground)',
        '--rv-search-action-text-color': 'var(--foreground)',
        '--rv-cell-text-color': 'var(--foreground)',
        '--rv-input-text-color': 'var(--foreground)',
        '--rv-input-value-color': 'var(--foreground)',
        '--rv-input-placeholder-color': 'var(--foreground)',
        'rv-text-color': 'var(--foreground)',
    };

    const spring = {
        type: 'spring',
        stiffness: 300,
        damping: 10,
        mass: 0.6,
        bounce: 0.5,
    };

    const containerVariants = {
        hidden: {opacity: 0},
        visible: {
            opacity: 1,
            transition: {
                staggerChildren: 0.05,
            },
        },
    };

    const itemVariants = {
        hidden: {x: -100, opacity: 0},
        visible: {x: 0, opacity: 1, transition: spring},
    };

    const debounce = useCallback((func: (...args: unknown[]) => void, wait: number) => {
        let timeout: NodeJS.Timeout;
        return (...args: unknown[]) => {
            clearTimeout(timeout);
            timeout = setTimeout(() => func(...args), wait);
        };
    }, []);

    const likeHandle = () => {
        likeRequest(type, targetId).then((res) => {
            if (res) {
                toast('点赞成功，感谢您的支持！');
            } else {
                toast('您已经点过赞了捏！感谢！', {type: 'info'});
            }
        });
    };

    const getDoms = useCallback((items: TocItem[]): HTMLElement[] => {
        const doms: HTMLElement[] = [];
        const addToDoms = (items: TocItem[]) => {
            for (const item of items) {
                const dom = document.getElementById(item.anchor);
                if (dom) {
                    doms.push(dom);
                }
                if (item.children && item.children.length) {
                    addToDoms(item.children);
                }
            }
        };
        if (typeof document !== 'undefined' && items.length) {
            addToDoms(items);
        }
        return doms;
    }, []);

    useEffect(() => {
        emitter.on('showTitle', () => {
            setIsUtilIconShowed(false);
        });
        emitter.on('hideTitle', () => {
            setIsUtilIconShowed(true);
        });
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-expect-error
        emitter.on('readingProgress', (progress: number) => {
            setIsTopIconShowed(progress > 5);
        });
    }, []);

    useEffect(() => {
        if (typeof document !== 'undefined') {
            setDoms(getDoms(toc));
        }
    }, [toc, getDoms]);

    useEffect(() => {
        const handleScroll = debounce(() => {
            const range = window.innerHeight;
            let newActiveAnchor = '';
            for (const dom of doms) {
                if (!dom) continue;
                const top = dom.getBoundingClientRect().top;
                if (top >= 0 && top <= range) {
                    newActiveAnchor = dom.id;
                    break;
                } else if (top > range) {
                    break;
                } else {
                    newActiveAnchor = dom.id;
                }
            }
            if (newActiveAnchor !== activeAnchor) {
                setActiveAnchor(newActiveAnchor);
            }
        }, 50);

        emitter.on('scroll', handleScroll);
        return () => {
            emitter.off('scroll', handleScroll);
        };
    }, [doms, activeAnchor, debounce]);

    const renderTocItems = (items: TocItem[]) => {
        return items.map((item, index) => (
            <motion.div
                key={index}
                style={{marginLeft: `${(item.level - 2) * 20}px`, margin: '5px'}}
                variants={itemVariants}
                layout
            >
                <li className={item.anchor === activeAnchor ? 'font-semibold text-primary' : 'font-normal'}>
                    <a href={`#${item.anchor}`}>{item.name}</a>
                </li>
                {item.children && item.children.length > 0 && (
                    <ul>
                        {renderTocItems(item.children)}
                    </ul>
                )}
            </motion.div>
        ));
    };

    return isMobile ? (
        <>
            {/* 从右下角模糊到左上角清晰的背景遮罩 */}
            {isUtilIconShowed && (
                <div
                    style={{
                        position: 'fixed',
                        bottom: 0,
                        right: 0,
                        width: '15rem',
                        height: isTopIconShowed ? '20rem' : '16rem',
                        backdropFilter: 'blur(4px) saturate(180%) contrast(120%)',
                        WebkitBackdropFilter: 'blur(4px) saturate(180%) contrast(120%)',
                        background: `
                            radial-gradient(ellipse 120% 100% at bottom right, 
                                rgba(var(--background), 0.9) 0%, 
                                rgba(var(--background), 0.8) 20%, 
                                rgba(var(--background), 0.6) 35%, 
                                rgba(var(--background), 0.4) 50%, 
                                rgba(var(--background), 0.2) 65%, 
                                rgba(var(--background), 0.1) 80%, 
                                transparent 100%
                            )
                        `,
                        maskImage: `
                            radial-gradient(ellipse 150% 120% at bottom right, 
                                black 0%, 
                                black 15%, 
                                rgba(0,0,0,0.9) 25%, 
                                rgba(0,0,0,0.7) 40%, 
                                rgba(0,0,0,0.5) 55%, 
                                rgba(0,0,0,0.3) 70%, 
                                rgba(0,0,0,0.1) 85%, 
                                transparent 100%
                            )
                        `,
                        WebkitMaskImage: `
                            radial-gradient(ellipse 150% 120% at bottom right, 
                                black 0%, 
                                black 15%, 
                                rgba(0,0,0,0.9) 25%, 
                                rgba(0,0,0,0.7) 40%, 
                                rgba(0,0,0,0.5) 55%, 
                                rgba(0,0,0,0.3) 70%, 
                                rgba(0,0,0,0.1) 85%, 
                                transparent 100%
                            )
                        `,
                        zIndex: 9,
                        transition: 'all 0.3s ease-in-out',
                        pointerEvents: 'none',
                    }}
                />
            )}

            {isTopIconShowed && <Button
                size="icon"
                className="rounded-full bg-background text-foreground"
                style={{
                    position: 'fixed',
                    bottom: '13rem',
                    right: isUtilIconShowed ? '1rem' : '-3rem',
                    transition: 'right 0.3s ease-in-out',
                    zIndex: 1000,
                    border: '1px solid rgba(var(--foreground), 0.1)',
                }}
                onClick={() => {
                    window.scrollTo({top: 0, behavior: 'smooth'})
                }}
            >
                <PinTopIcon className="h-6 w-6"/>
            </Button>
            }
            <Button
                size="icon"
                className="rounded-full bg-background text-foreground"
                style={{
                    position: 'fixed',
                    bottom: '10rem',
                    right: isUtilIconShowed ? '1rem' : '-3rem',
                    transition: 'right 0.3s ease-in-out',
                    zIndex: 1000,
                    border: '1px solid rgba(var(--foreground), 0.1)',
                }}
                onClick={likeHandle}
            >
                <HeartIcon className="h-6 w-6"/>
            </Button>
            <Button
                size="icon"
                className="rounded-full bg-background text-foreground"
                style={{
                    position: 'fixed',
                    bottom: '7rem',
                    right: isUtilIconShowed ? '1rem' : '-3rem',
                    transition: 'right 0.3s ease-in-out',
                    zIndex: 1000,
                    border: '1px solid rgba(var(--foreground), 0.1)',
                }}
                onClick={() => {
                    document.scrollingElement?.scrollTo({
                        top: document.scrollingElement.scrollHeight - window.innerHeight - 600,
                        behavior: 'smooth'
                    });
                }}
            >
                <MessageCircleIcon className="h-6 w-6"/>
            </Button>
            <Button
                size="icon"
                className="rounded-full bg-background text-foreground"
                style={{
                    position: 'fixed',
                    bottom: '4rem',
                    right: isUtilIconShowed ? '1rem' : '-3rem',
                    transition: 'right 0.3s ease-in-out',
                    zIndex: 1000,
                    border: '1px solid rgba(var(--foreground), 0.1)',
                }}
                onClick={showPanelHandle}
            >
                <TableOfContentsIcon className="h-6 w-6"/>
            </Button>
            <AnimatePresence>
                {showPanel && (
                    <>
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            transition={{ duration: 0.2 }}
                            style={{
                                position: 'fixed',
                                overflow: 'hidden',
                                top: 0,
                                left: 0,
                                width: '100%',
                                height: '100%',
                                zIndex: 1000,
                            }}
                            onClick={hidePanelHandle}
                        />
                        <motion.div
                            initial={{ y: '100%' }}
                            animate={{ y: 0 }}
                            exit={{ y: '100%' }}
                            transition={{ 
                                type: 'spring',
                                stiffness: 300,
                                damping: 30
                            }}
                            style={{
                                position: 'fixed',
                                bottom: 0,
                                left: 0,
                                right: 0,
                                zIndex: 1001,
                                backgroundColor: 'rgba(var(--background), 1)',
                                borderTopLeftRadius: '20px',
                                borderTopRightRadius: '20px',
                                maxHeight: '80vh',
                                overflow: 'hidden',
                                backdropFilter: 'blur(10px)',
                                WebkitBackdropFilter: 'blur(10px)',
                                border: '1px solid rgba(var(--foreground), 0.1)',
                                borderBottom: 'none',
                            }}
                        >
                            <div style={{ 
                                display: 'flex',
                                justifyContent: 'center',
                                padding: '8px',
                                borderBottom: '1px solid rgba(var(--foreground), 0.1)'
                            }}>
                                <div style={{
                                    width: '20px',
                                    height: '4px',
                                    backgroundColor: 'rgba(var(--foreground), 0.3)',
                                    borderRadius: '2px'
                                }}></div>
                            </div>
                            <ConfigProvider themeVars={themeVars}>
                                <Search style={{ padding: '10px' }}/>
                            </ConfigProvider>
                            <motion.ul
                                style={{
                                    overflow: 'auto',
                                    padding: '20px',
                                    maxHeight: 'calc(80vh - 120px)',
                                }}
                                initial="hidden"
                                animate="visible"
                                variants={containerVariants}
                            >
                                {toc.length > 0 && renderTocItems(toc)}
                            </motion.ul>
                        </motion.div>
                    </>
                )}
            </AnimatePresence>
        </>
    ) : null;
};

export default FloatingTocMobile;
