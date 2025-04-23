import React from 'react';

interface SpinnerProps {
  size?: 'small' | 'medium' | 'large';
  message?: string;
}

const Spinner: React.FC<SpinnerProps> = ({ size = 'medium', message = 'Loading...' }) => {
  const sizeClasses = {
    small: 'w-6 h-6 border-2',
    medium: 'w-10 h-10 border-3',
    large: 'w-16 h-16 border-4'
  };

  return (
    <div className="flex flex-col items-center justify-center py-6">
      <div className={`${sizeClasses[size]} border-t-blue-500 border-blue-200 rounded-full animate-spin`}></div>
      {message && <p className="mt-3 text-gray-600">{message}</p>}
    </div>
  );
};

export default Spinner;