import { useState, useCallback } from 'react';

interface JsonBrowserProps {
  data: any[];
}

const JsonBrowser: React.FC<JsonBrowserProps> = ({ data }) => {
  const [expanded, setExpanded] = useState<Set<number>>(new Set());

  const toggleItem = useCallback((index: number) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(index)) {
        next.delete(index);
      } else {
        next.add(index);
      }
      return next;
    });
  }, []);

  const allExpanded = data.length > 0 && expanded.size === data.length;

  const toggleAll = useCallback(() => {
    if (allExpanded) {
      setExpanded(new Set());
    } else {
      setExpanded(new Set(data.map((_, i) => i)));
    }
  }, [allExpanded, data]);

  const summarize = (item: any): string => {
    if (item._raw) return item._raw;
    const keys = Object.keys(item);
    if (keys.length === 0) return '{}';
    const preview = keys.slice(0, 3).map((k) => `${k}: ${JSON.stringify(item[k])}`).join(', ');
    return `{ ${preview}${keys.length > 3 ? ', ...' : ''} }`;
  };

  if (data.length === 0) {
    return (
      <div className="text-term-text-muted font-mono text-xs py-2">
        No output yet. Execute the configuration to see results.
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-end mb-1">
        <button
          onClick={toggleAll}
          className="text-term-text-muted hover:text-term-text p-1"
          title={allExpanded ? 'Collapse all' : 'Expand all'}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 20 20"
            fill="currentColor"
            className="w-4 h-4"
          >
            {allExpanded ? (
              <path
                fillRule="evenodd"
                d="M3.28 2.22a.75.75 0 00-1.06 1.06L5.44 6.5H2.75a.75.75 0 000 1.5h4.5a.75.75 0 00.75-.75v-4.5a.75.75 0 00-1.5 0v2.69L3.28 2.22zM16.72 2.22a.75.75 0 010 1.06L13.56 6.5h2.69a.75.75 0 010 1.5h-4.5a.75.75 0 01-.75-.75v-4.5a.75.75 0 011.5 0v2.69l3.22-3.22a.75.75 0 011.06 0zM3.28 17.78a.75.75 0 001.06 0L7.56 14.5H4.87a.75.75 0 010-1.5h4.5a.75.75 0 01.75.75v4.5a.75.75 0 01-1.5 0v-2.69l-3.22 3.22a.75.75 0 01-1.06-1.06h-.06zM16.72 17.78a.75.75 0 01-1.06 0L12.44 14.5h2.69a.75.75 0 000-1.5h-4.5a.75.75 0 00-.75.75v4.5a.75.75 0 001.5 0v-2.69l3.22 3.22a.75.75 0 001.06-1.06h.06z"
                clipRule="evenodd"
              />
            ) : (
              <path
                fillRule="evenodd"
                d="M13.28 7.78l3.22-3.22v2.69a.75.75 0 001.5 0v-4.5a.75.75 0 00-.75-.75h-4.5a.75.75 0 000 1.5h2.69l-3.22 3.22a.75.75 0 001.06 1.06zM2 17.25v-4.5a.75.75 0 011.5 0v2.69l3.22-3.22a.75.75 0 011.06 1.06L4.56 16.5h2.69a.75.75 0 010 1.5h-4.5a.75.75 0 01-.75-.75zM12.22 13.28l3.22 3.22h-2.69a.75.75 0 000 1.5h4.5a.75.75 0 00.75-.75v-4.5a.75.75 0 00-1.5 0v2.69l-3.22-3.22a.75.75 0 00-1.06 1.06zM7.78 6.72L4.56 3.5h2.69a.75.75 0 000-1.5h-4.5a.75.75 0 00-.75.75v4.5a.75.75 0 001.5 0V4.56l3.22 3.22a.75.75 0 001.06-1.06z"
                clipRule="evenodd"
              />
            )}
          </svg>
        </button>
      </div>
      <div className="space-y-px">
        {data.map((item, index) => {
          const isExpanded = expanded.has(index);
          return (
            <div key={index}>
              <button
                onClick={() => toggleItem(index)}
                className="w-full flex items-center gap-2 px-2 py-1 text-left hover:bg-term-bg-muted rounded transition-colors group"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 16 16"
                  fill="currentColor"
                  className={`w-3 h-3 text-term-text-muted shrink-0 transition-transform ${isExpanded ? 'rotate-90' : ''}`}
                >
                  <path d="M6.22 4.22a.75.75 0 0 1 1.06 0l3.25 3.25a.75.75 0 0 1 0 1.06l-3.25 3.25a.75.75 0 0 1-1.06-1.06L8.94 8 6.22 5.28a.75.75 0 0 1 0-1.06Z" />
                </svg>
                <span className="text-term-cyan text-xs font-mono">[{index}]</span>
                {!isExpanded && (
                  <span className="text-term-text-muted text-xs font-mono truncate">
                    {summarize(item)}
                  </span>
                )}
              </button>
              {isExpanded && (
                <pre className="ml-7 px-2 py-1 text-xs font-mono text-term-text overflow-x-auto">
                  {JSON.stringify(item, null, 2)}
                </pre>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default JsonBrowser;
