import { RefreshCw } from "lucide-react";

interface LoadingSpinnerProps {
  message?: string;
  size?: "sm" | "md" | "lg";
}

export const LoadingSpinner = ({
  message = "Loading...",
  size = "md",
}: LoadingSpinnerProps) => {
  const sizeClasses = {
    sm: "w-4 h-4",
    md: "w-6 h-6",
    lg: "w-8 h-8",
  };

  return (
    <div className="flex items-center justify-center p-8">
      <div className="flex items-center space-x-2">
        <RefreshCw
          className={`${sizeClasses[size]} animate-spin text-blue-500`}
        />
        <span className="text-gray-600 dark:text-gray-400">{message}</span>
      </div>
    </div>
  );
};
