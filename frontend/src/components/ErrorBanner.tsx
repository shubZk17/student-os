import React from 'react';
import { AlertCircle } from 'lucide-react';

export const ErrorBanner: React.FC<{ message: string; onDismiss?: () => void }> = ({ message, onDismiss }) => (
  <div role="alert" className="flex items-center justify-between gap-3 p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs">
    <div className="flex items-center gap-2">
      <AlertCircle className="h-4 w-4 shrink-0" />
      <span>{message}</span>
    </div>
    {onDismiss && (
      <button onClick={onDismiss} className="font-semibold hover:text-white">
        Dismiss
      </button>
    )}
  </div>
);
