'use client'

import * as React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { ScrollArea, Theme } from '@radix-ui/themes';
import CommentArea from "@/components/comment/CommentArea";

interface CommentModalProps {
    isOpen: boolean
    onClose: () => void
    commentId: string
}

function CommentModal({isOpen = false, onClose, commentId}: CommentModalProps) {
    return (
        <Dialog open={isOpen} onOpenChange={onClose}>
            <DialogContent 
                className="max-w-4xl max-h-[80vh] p-6"
                style={{
                    background: 'rgba(var(--background), 0.8)',
                    backdropFilter: 'blur(50px)',
                    outline: '1px solid rgba(var(--foreground), 0.1)',
                    borderRadius: '5px',
                }}
            >
                <Theme>
                    <DialogHeader className="mb-4">
                        <DialogTitle>评论</DialogTitle>
                    </DialogHeader>
                    <ScrollArea style={{ maxHeight: 'calc(80vh - 120px)' }}>
                        <CommentArea id={commentId} isModal={true}/>
                    </ScrollArea>
                </Theme>
            </DialogContent>
        </Dialog>
    )
}

export default CommentModal;
