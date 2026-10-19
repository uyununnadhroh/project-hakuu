const ErrorBanner = ({ message }) => {
    if (!message) return null;
    return (
        <div className="bg-red-50 border border-red-200 rounded-xl px-4 py-3 text-sm text-[#b74b40]">
            {message}
        </div>
    );
};

export default ErrorBanner;
