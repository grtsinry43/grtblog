import React from 'react';
import {Avatar, Button} from '@radix-ui/themes';
import Image from 'next/image';
import {ArrowRight, Eye, Heart, MessageCircle, Share2} from 'lucide-react';
import Link from 'next/link';
import {motion} from 'framer-motion';
import {Moment} from '@/types';
import {formatDistanceToNow, parseISO} from 'date-fns';
import {zhCN} from 'date-fns/locale';
import {PinTopIcon} from '@radix-ui/react-icons';
import {clsx} from 'clsx';
import {article_font} from '@/app/fonts/font';
import CommentModal from "@/components/comment/CommentModal";

const MomentCardItem = ({moment, index}: { moment: Moment, index: number }) => {
    const formattedCreatedDate = formatDistanceToNow(parseISO(moment.createdAt), {addSuffix: true, locale: zhCN});
    const [isCommentModalOpen, setIsCommentModalOpen] = React.useState(false);

    const onClose = () => {
        setIsCommentModalOpen(false);
    }

    // 判断是否有图片
    const hasImages = moment.images && moment.images.length > 0 && moment.images[0] !== '';

    return (
        <div>
            <CommentModal isOpen={isCommentModalOpen} onClose={onClose} commentId={moment.commentId}/>
            <motion.div
                key={moment.shortUrl}
                initial={{opacity: 0, y: 50}}
                animate={{opacity: 1, y: 0}}
                transition={{duration: 0.5, delay: index * 0.1}}
            >
                <div className={clsx(article_font.className, 'mb-8 flex justify-between items-center w-full relative')}>
                    {/* 精美的卡片容器 */}
                    <div className="w-full z-10 bg-white dark:bg-gray-900/95 backdrop-blur-sm p-6 sm:p-8 relative overflow-hidden group hover:shadow-xl transition-all duration-300" 
                         style={{
                             borderRadius: '1rem',
                             border: '1px solid rgba(var(--foreground), 0.08)',
                             boxShadow: '0 4px 20px rgba(0, 0, 0, 0.05), 0 1px 3px rgba(0, 0, 0, 0.1)',
                         }}>
                        
                        {/* 顶部渐变装饰 */}
                        <div className="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 opacity-60"/>
                        
                        {/* 卡片内容 */}
                        <div className="relative">
                            {/* 作者信息区域 */}
                            <div className="flex flex-row items-center gap-3 mb-4">
                                <div className="relative">
                                    <Avatar src={moment.authorAvatar} alt={moment.authorName}
                                            fallback={moment.authorName[0]}
                                            radius={"full"}
                                            size="3"
                                    />
                                    <div className="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-green-500 rounded-full border-2 border-white dark:border-gray-900"/>
                                </div>
                                <div className="flex-grow">
                                    <div className="font-medium text-base text-gray-900 dark:text-gray-100">{moment.authorName}</div>
                                    <p className="text-xs text-gray-500 dark:text-gray-400">{formattedCreatedDate}</p>
                                </div>
                                <div className="flex gap-2 items-center">
                                    {moment.isTop && (
                                        <div className="flex items-center justify-center w-6 h-6 bg-gradient-to-r from-orange-400 to-red-500 rounded-full">
                                            <PinTopIcon className="text-white w-3 h-3"/>
                                        </div>
                                    )}
                                    {moment.isHot && (
                                        <span className="bg-gradient-to-r from-red-500 to-pink-500 text-white text-xs px-2 py-1 rounded-full font-medium shadow-sm">
                                            HOT
                                        </span>
                                    )}
                                </div>
                            </div>

                            {/* 标题和内容 */}
                            <div className="space-y-3">
                                <h3 className="text-lg sm:text-xl font-semibold text-gray-900 dark:text-gray-100 leading-snug">{moment.title}</h3>
                                <p className="text-gray-600 dark:text-gray-300 text-sm leading-relaxed">{moment.summary}</p>
                                
                                {/* 条件渲染图片区域 */}
                                {hasImages && (
                                    <div className={clsx(
                                        "mt-4 rounded-lg overflow-hidden",
                                        moment.images.length === 1 ? "grid grid-cols-1" : "grid grid-cols-2 gap-2"
                                    )}>
                                        {moment.images.slice(0, 4).map((image, imgIndex) => (
                                            <motion.div 
                                                key={imgIndex} 
                                                className="relative aspect-video bg-gray-100 dark:bg-gray-800 rounded-lg overflow-hidden group"
                                                whileHover={{ scale: 1.02 }}
                                                transition={{ duration: 0.2 }}
                                            >
                                                <Image
                                                    src={image.startsWith('http') ? image : `${process.env.NEXT_PUBLIC_BASE_URL}/${image.slice(1)}`}
                                                    alt={`Image ${imgIndex + 1} for ${moment.title}`}
                                                    fill
                                                    className="object-cover transition-transform duration-300 group-hover:scale-105"
                                                />
                                                <div className="absolute inset-0 bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity duration-300"/>
                                            </motion.div>
                                        ))}
                                        {moment.images.length > 4 && (
                                            <div className="relative aspect-video bg-gray-100 dark:bg-gray-800 rounded-lg overflow-hidden flex items-center justify-center">
                                                <span className="text-gray-500 text-sm font-medium">+{moment.images.length - 4}</span>
                                            </div>
                                        )}
                                    </div>
                                )}
                            </div>

                            {/* 底部操作区域 */}
                            <div className="flex justify-between items-center mt-6 pt-4 border-t border-gray-100 dark:border-gray-800">
                                <div className="flex gap-6">
                                    <motion.span 
                                        className="flex items-center gap-1.5 text-gray-500 dark:text-gray-400 text-sm cursor-pointer hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
                                        whileHover={{ scale: 1.05 }}
                                    >
                                        <Eye size={14}/> 
                                        <span className="font-medium">{moment.views}</span>
                                    </motion.span>
                                    <motion.span 
                                        className="flex items-center gap-1.5 text-gray-500 dark:text-gray-400 text-sm cursor-pointer hover:text-red-500 dark:hover:text-red-400 transition-colors"
                                        whileHover={{ scale: 1.05 }}
                                    >
                                        <Heart size={14}/> 
                                        <span className="font-medium">{moment.likes}</span>
                                    </motion.span>
                                    <motion.span 
                                        className="flex items-center gap-1.5 text-gray-500 dark:text-gray-400 text-sm cursor-pointer hover:text-green-600 dark:hover:text-green-400 transition-colors"
                                        onClick={() => setIsCommentModalOpen(true)}
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                    >
                                        <MessageCircle size={14}/> 
                                        <span className="font-medium">{moment.comments}</span>
                                    </motion.span>
                                    <motion.span 
                                        className="flex items-center gap-1.5 text-gray-500 dark:text-gray-400 text-sm cursor-pointer hover:text-purple-600 dark:hover:text-purple-400 transition-colors"
                                        whileHover={{ scale: 1.05 }}
                                        whileTap={{ scale: 0.95 }}
                                    >
                                        <Share2 size={14}/>
                                    </motion.span>
                                </div>
                                <Link href={`/moments/${moment.shortUrl}`} passHref>
                                    <motion.div
                                        whileHover={{ scale: 1.02 }}
                                        whileTap={{ scale: 0.98 }}
                                    >
                                        <Button 
                                            variant="ghost" 
                                            size="2" 
                                            className="text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 font-medium text-sm px-4 py-2 rounded-lg transition-all duration-200"
                                        >
                                            查看详情 
                                            <ArrowRight className="ml-1.5 h-3.5 w-3.5"/>
                                        </Button>
                                    </motion.div>
                                </Link>
                            </div>
                        </div>
                    </div>
                    
                    {/* 时间轴圆点 */}
                    <div className="absolute left-1/2 w-4 h-4 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full transform -translate-x-1/2 border-2 border-white dark:border-gray-900 shadow-lg"/>
                </div>
            </motion.div>
        </div>
    );
};

export default MomentCardItem;
