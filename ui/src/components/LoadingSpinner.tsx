const LoadingSpinner = () => {
  return (
    <div className="flex justify-center items-center h-64">
      <svg
        className="animate-spin h-16 w-16 text-blue-800"
        role="status"
        aria-label="Loading"
        viewBox="0 0 24 24"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        {/* Detailed goat body */}
        <path
          d="M12 2L10.5 3L9.5 4.5L9 6L8.5 7.5L8 9L7.5 10L7 11L6.5 12L7 13L7.5 14L8 15L9 16L10 17L11 17.5L12 18L13 17.5L14 17L15 16L16 15L16.5 14L17 13L17.5 12L17 11L16.5 10L16 9L15.5 7.5L15 6L14.5 4.5L13.5 3L12 2Z"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
        {/* Goat horns */}
        <path
          d="M10.5 4L9 3L8 4L7.5 5L8 6L9 6.5L10 6L10.5 5L10.5 4Z M13.5 4L15 3L16 4L16.5 5L16 6L15 6.5L14 6L13.5 5L13.5 4Z"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
        {/* Goat face details */}
        <path
          d="M10 8L9.5 9L10 10L11 10.5L12 10L13 10.5L14 10L14.5 9L14 8L13 7.5L12 8L11 7.5L10 8Z"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
        {/* Goat eyes */}
        <circle cx="10.5" cy="9" r="0.5" fill="currentColor" />
        <circle cx="13.5" cy="9" r="0.5" fill="currentColor" />
        {/* Goat beard */}
        <path
          d="M11.5 10.5L12 11L12.5 10.5L12 12L11.5 13L12 14L12.5 13L12 12"
          stroke="currentColor"
          strokeWidth="1"
          strokeLinecap="round"
        />
        {/* Goat legs */}
        <path
          d="M9 16L8 18L7.5 20L8 21L8.5 20L9 18 M15 16L16 18L16.5 20L16 21L15.5 20L15 18"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
      </svg>
    </div>
  );
};

export default LoadingSpinner; 