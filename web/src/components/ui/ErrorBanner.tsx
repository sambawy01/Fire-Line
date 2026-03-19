import { AlertCircle, X, RefreshCw } from 'lucide-react';

interface ErrorBannerProps {
  message: string;
  onDismiss?: () => void;
  retry?: () => void;
}

export default function ErrorBanner({ message, onDismiss, retry }: ErrorBannerProps) {
  return (
    <div className="rounded-lg border border-red-200 bg-red-50 p-4 flex items-start gap-3">
      <AlertCircle className="h-5 w-5 text-red-500 shrink-0 mt-0.5" />
      <div className="flex-1 min-w-0">
        <p className="text-sm text-red-700">{message}</p>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        {retry && (
          <button onClick={retry} className="text-red-500 hover:text-red-700 transition-colors">
            <RefreshCw className="h-4 w-4" />
          </button>
        )}
        {onDismiss && (
          <button onClick={onDismiss} className="text-red-400 hover:text-red-600 transition-colors">
            <X className="h-4 w-4" />
          </button>
        )}
      </div>
    </div>
  );
}
