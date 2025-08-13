import { RefreshCw, WifiOff } from "lucide-react";

interface ErrorMessageProps {
  message: string;
  onRetry?: () => void;
  compact?: boolean;
}

export const ErrorMessage = ({
  message,
  onRetry,
  compact = false,
}: ErrorMessageProps) => {
  if (compact) {
    return (
      <div className="flex items-center justify-between p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
        <div className="flex items-center space-x-2">
          <WifiOff className="w-4 h-4 text-red-500" />
          <span className="text-sm text-red-700 dark:text-red-300">
            {message}
          </span>
        </div>
        {onRetry && (
          <button
            onClick={onRetry}
            className="text-xs px-2 py-1 bg-red-100 dark:bg-red-800 text-red-700 dark:text-red-200 rounded hover:bg-red-200 dark:hover:bg-red-700 transition-colors"
          >
            Retry
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center p-8">
      <div className="text-center">
        <WifiOff className="w-12 h-12 text-red-500 mx-auto mb-4" />
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          Connection Error
        </h3>
        <p className="text-gray-600 dark:text-gray-400 mb-4">{message}</p>
        {onRetry && (
          <button
            onClick={onRetry}
            className="inline-flex items-center space-x-2 px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 transition-colors"
          >
            <RefreshCw className="w-4 h-4" />
            <span>Try Again</span>
          </button>
        )}
      </div>
    </div>
  );
};
