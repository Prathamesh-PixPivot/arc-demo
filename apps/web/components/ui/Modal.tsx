'use client';

import { useEffect, useRef, HTMLAttributes } from 'react';
import { createPortal } from 'react-dom';
import { cn } from '@/lib/utils';
import { Button } from './Button';

interface ModalProps extends HTMLAttributes<HTMLDivElement> {
    isOpen: boolean;
    onClose: () => void;
    title?: string;
    description?: string;
    footer?: React.ReactNode;
}

export function Modal({
    isOpen,
    onClose,
    title,
    description,
    children,
    footer,
    className,
    ...props
}: ModalProps) {
    const overlayRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape') onClose();
        };

        if (isOpen) {
            document.addEventListener('keydown', handleEscape);
            document.body.style.overflow = 'hidden';
        }

        return () => {
            document.removeEventListener('keydown', handleEscape);
            document.body.style.overflow = 'unset';
        };
    }, [isOpen, onClose]);

    if (!isOpen) return null;

    return createPortal(
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            {/* Backdrop */}
            <div
                ref={overlayRef}
                className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity animate-fade-in"
                onClick={onClose}
                aria-hidden="true"
            />

            {/* Modal Content */}
            <div
                className={cn(
                    'relative z-50 w-full max-w-lg transform rounded-lg bg-white p-6 shadow-xl transition-all animate-slide-up sm:mx-auto',
                    className
                )}
                role="dialog"
                aria-modal="true"
                aria-labelledby={title ? 'modal-title' : undefined}
                aria-describedby={description ? 'modal-description' : undefined}
                {...props}
            >
                <div className="flex items-center justify-between mb-4">
                    {title && (
                        <h2 id="modal-title" className="text-lg font-semibold text-gray-900">
                            {title}
                        </h2>
                    )}
                    <button
                        onClick={onClose}
                        className="rounded-full p-1 hover:bg-gray-100 transition-colors focus-ring"
                        aria-label="Close modal"
                    >
                        <svg
                            className="w-5 h-5 text-gray-500"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={2}
                            stroke="currentColor"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                {description && (
                    <p id="modal-description" className="text-sm text-gray-500 mb-4">
                        {description}
                    </p>
                )}

                <div className="mt-2">{children}</div>

                {footer && <div className="mt-6 flex justify-end gap-3">{footer}</div>}
            </div>
        </div>,
        document.body
    );
}
