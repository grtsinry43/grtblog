export const formatRelativeTime = (dateStr: string): string => {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();

    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days < 1) {
        if (hours < 1) {
            if (minutes < 1) return '刚刚';
            return `${minutes} 分钟前`;
        }
        return `${hours} 小时前`;
    }

    if (days < 7) return `${days} 天前`;
    if (days < 30) return `大约 ${Math.ceil(days / 7)} 周前`;
    if (days < 365) return `大约 ${Math.floor(days / 30)} 个月前`;

    return `${date.getFullYear()}年`;
};
