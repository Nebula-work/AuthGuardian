import React, { useState, useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Components } from 'react-markdown';
import { useParams, Link, useNavigate } from 'react-router-dom';

// Define the documentation files available
const docFiles = [
  { id: 'api', title: 'API Reference', file: 'API.md' },
  { id: 'architecture', title: 'Architecture Guide', file: 'ARCHITECTURE.md' },
  { id: 'integration', title: 'Integration Guide', file: 'INTEGRATION.md' },
  { id: 'contributing', title: 'Contributing Guide', file: 'CONTRIBUTING.md' }
];

const Documentation: React.FC = () => {
  const { docId = 'api' } = useParams<{ docId?: string }>();
  const [content, setContent] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchDocumentation = async () => {
      try {
        setLoading(true);
        setError(null);

        // Find the documentation file based on the ID parameter
        const docFile = docFiles.find(doc => doc.id === docId);
        
        if (!docFile) {
          setError(`Documentation "${docId}" not found`);
          setLoading(false);
          return;
        }

        // Fetch the markdown file from the server
        const response = await fetch(`/docs/${docFile.file}`);
        
        if (!response.ok) {
          throw new Error(`Failed to load documentation: ${response.statusText}`);
        }
        
        const text = await response.text();
        setContent(text);
      } catch (err) {
        console.error('Error fetching documentation:', err);
        setError('Failed to load documentation. Please try again later.');
      } finally {
        setLoading(false);
      }
    };

    fetchDocumentation();
  }, [docId]);

  // Handle selecting a different documentation file
  const handleDocSelect = (id: string) => {
    navigate(`/documentation/${id}`);
  };

  return (
    <div className="flex flex-col p-6">
      <h1 className="text-3xl font-bold mb-6">Documentation</h1>
      
      <div className="flex flex-col md:flex-row gap-6">
        {/* Documentation Navigation Sidebar */}
        <div className="w-full md:w-64 flex-shrink-0">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <h2 className="text-xl font-semibold mb-4">Documentation Files</h2>
            <ul className="space-y-2">
              {docFiles.map((doc) => (
                <li key={doc.id}>
                  <button
                    onClick={() => handleDocSelect(doc.id)}
                    className={`w-full text-left px-3 py-2 rounded-md transition-colors ${
                      docId === doc.id
                        ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-200'
                        : 'hover:bg-gray-100 dark:hover:bg-gray-700'
                    }`}
                  >
                    {doc.title}
                  </button>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Documentation Content */}
        <div className="flex-grow bg-white dark:bg-gray-800 rounded-lg shadow p-6">
          {loading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
            </div>
          ) : error ? (
            <div className="bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-200 p-4 rounded-md">
              {error}
            </div>
          ) : (
            <div className="documentation-content prose dark:prose-invert max-w-none">
              {/* @ts-ignore - Type issues with ReactMarkdown */}
              <ReactMarkdown
                components={{
                  code: ({ node, inline, className, children, ...props }: any) => {
                    const match = /language-(\w+)/.exec(className || '');
                    return !inline && match ? (
                      <SyntaxHighlighter
                        language={match[1]}
                        style={vscDarkPlus}
                        PreTag="div"
                        {...props}
                      >
                        {String(children).replace(/\n$/, '')}
                      </SyntaxHighlighter>
                    ) : (
                      <code className={className} {...props}>
                        {children}
                      </code>
                    );
                  },
                }}
              >
                {content}
              </ReactMarkdown>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Documentation;