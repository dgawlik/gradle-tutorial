import { useState, useEffect } from 'react';
import './App.css';
import InterleaveView from './components/InterleaveView';
import Spinner from './components/Spinner';

// Types matching our Go structs
interface TranslationEntry {
  id: number;
  originalText: string;
  translation: string;
  language: string;
  createdAt: string;
}

function App() {
  const [view, setView] = useState<'home' | 'translate' | 'history' | 'detail' | 'interleave'>('home');
  const [translations, setTranslations] = useState<TranslationEntry[]>([]);
  const [currentTranslation, setCurrentTranslation] = useState<TranslationEntry | null>(null);
  const [text, setText] = useState('');
  const [language, setLanguage] = useState('German');
  const [hoveredWord, setHoveredWord] = useState<string | null>(null);
  const [wordDefinitions, setWordDefinitions] = useState<string[]>([]);
  const [ttsInfo, setTtsInfo] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);

  // Load translations on startup
  useEffect(() => {
    loadTranslations();
  }, []);

  const loadTranslations = async () => {
    const translations: TranslationEntry[] = await fetch('/api/translations').then(response => response.json())
    setTranslations(translations)
  };

  const handleTranslate = async () => {
    if (!text.trim()) {
      alert('Please enter text to translate.');
      return;
    }

    setIsLoading(true);
    try {
      const response = await fetch('/api/newtranslation', {
        method: 'POST',
        body: text
      });

      if (!response.ok) {
        throw new Error('Translation failed');
      }

      const translation = await response.json();
      setCurrentTranslation(translation);
      setText('');
      loadTranslations();
    } catch (error) {
      console.error('Error during translation:', error);
    } finally {
      setIsLoading(false);
    }
  };



  const viewTranslation = async (id: number) => {
    const translation: TranslationEntry = await fetch('/api/translations/' + id).then(response => response.json());
    setCurrentTranslation(translation);
    setView('detail')
  };

  const deleteTranslation = async (id: number) => {
     const response = await fetch('/api/translations/' + id, {
      method: 'DELETE'
    });
    if (!response.ok) {
      console.log('Error deleting translation');
    } else {
      loadTranslations();
      setView('history');
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 text-gray-900">
      <div className="container mx-auto px-4 py-8">
        <header className="mb-8">
          <h1 className="text-3xl font-bold text-center text-gray-900 mb-6">Language Learning App</h1>
          <nav className="flex justify-center space-x-10 border-b border-gray-200 pb-3">
            <span 
              onClick={() => setView('home')}
              className={`cursor-pointer ${view === 'home' ? 'font-semibold text-gray-900 border-b-2 border-gray-800' : 'text-gray-600 hover:text-gray-900'}`}
            >
              Home
            </span>
            <span>&nbsp;|&nbsp;</span>
            <span 
              data-testid="translate-btn1"
              onClick={() => setView('translate')}
              className={`cursor-pointer ${view === 'translate' ? 'font-semibold text-gray-900 border-b-2 border-gray-800' : 'text-gray-600 hover:text-gray-900'}`}
            >
              Translate
            </span>
            <span>&nbsp;|&nbsp;</span>
            <span 
              onClick={() => { setView('history'); loadTranslations(); }}
              className={`cursor-pointer ${view === 'history' ? 'font-semibold text-gray-900 border-b-2 border-gray-800' : 'text-gray-600 hover:text-gray-900'}`}
            >
              History
            </span>
            <span>&nbsp;|&nbsp;</span>
          </nav>
        </header>

        {view === 'home' && (
          <div className="text-center py-8  mx-auto">
            <h2 className="text-2xl mb-4 text-gray-900">Welcome to Language Learning</h2>
            <p className="mb-6 text-gray-700">This app helps you learn languages by translating texts and providing word-by-word translations.</p>
            <div className="flex flex-row justify-center items-center mt-8">
              <span 
                onClick={() => setView('translate')}
                className="cursor-pointer text-gray-900 font-medium hover:underline"
              >
                Start Translating
              </span>
              <span>&nbsp;|&nbsp;</span>
              <span 
                onClick={() => { setView('history'); loadTranslations(); }}
                className="cursor-pointer text-gray-700 hover:text-gray-900 hover:underline"
              >
                View History
              </span>
            </div>
          </div>
        )}

        {view === 'translate' && (
          <div className="bg-white p-6 rounded-lg shadow-sm w-4/5 mx-auto">
            <h2 className="text-2xl mb-4 text-gray-900">Translate Text</h2>
            <div className="mb-4">
              <label className="block text-gray-800 mb-2">Language:</label>
              <select 
                data-testid="language-select"
                value={language} 
                onChange={(e) => setLanguage(e.target.value)}
                className="w-1/4 p-2 border border-gray-300 rounded focus:outline-none focus:border-gray-500"
                disabled={isLoading}
              >
                <option value="German">German</option>
              </select>
            </div>
            <div className="mb-6 mt-6">
              <label className="block text-gray-800 mb-2">Text to translate:</label>
              <textarea 
                value={text}
                onChange={(e) => setText(e.target.value)}
                className="w-full p-3 border border-gray-300 rounded focus:outline-none focus:border-gray-500 h-40"
                placeholder="Enter text here..."
                disabled={isLoading}
              />
            </div>
            
            {!isLoading ? (
              <span 
                data-testid="translate-btn2"
                onClick={handleTranslate}
                className={`cursor-pointer ${!text.trim() ? 'text-gray-400' : 'text-gray-900 hover:underline'}`}
              >
                Translate
              </span>
            ) : (
              <Spinner size="medium" message="Translating text..." />
            )}
            
            {!isLoading && currentTranslation && (
              <div className="mt-8 border-t pt-6">
                <h3 className="text-xl font-semibold text-gray-800 mb-4">Translation Result</h3>
                <InterleaveView 
                  originalText={currentTranslation.originalText}
                  translatedText={currentTranslation.translation}
                  translationId={currentTranslation.id}
                />
              </div>
            )}
          </div>
        )}

        {view === 'history' && (
          <div className="bg-white p-6 rounded-lg shadow-sm w-4/5 mx-auto">
            <h2 className="text-2xl mb-4 text-gray-900">Translation History</h2>
            {!translations || translations.length === 0 ? (
              <p className="text-gray-600 text-center py-4">No translations yet.</p>
            ) : (
              <div className="space-y-4">
                {translations.map((item) => (
                  <div 
                    key={item.id} 
                    className="border border-gray-200 p-4 rounded hover:bg-gray-50" 
                  >
                    <p className="text-sm text-gray-600">{new Date(item.createdAt).toLocaleString()}</p>
                    <p className="font-medium my-1 text-gray-800">Language: {item.language}</p>
                    <p className="text-gray-700 truncate">{item.originalText.substring(0, 100)}...</p>
                    <div className="mt-3 pt-2 border-t border-gray-100 flex space-x-4">
                      <span
                        onClick={() => viewTranslation(item.id)}
                        className="cursor-pointer text-blue-600 hover:underline text-sm"
                      >
                        View Details
                      </span>
                      <span>&nbsp;|&nbsp;</span>
                      <span
                        onClick={() => deleteTranslation(item.id)}
                        className="cursor-pointer text-blue-600 hover:underline text-sm"
                      >
                        Delete
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {view === 'detail' && currentTranslation && (
          <div className="bg-white p-6 rounded-lg shadow-sm w-4/5 mx-auto">
            <h2 className="text-2xl mb-4 text-gray-900">Translation Detail</h2>
            
            <InterleaveView 
                  originalText={currentTranslation.originalText}
                  translatedText={currentTranslation.translation}
                  translationId={currentTranslation.id}
                />
            
            <div className="flex space-x-4">
              <span 
                onClick={() => setView('history')}
                className="cursor-pointer text-gray-700 hover:text-gray-900 hover:underline"
              >
                Back to History
              </span>
            </div>
          </div>
        )}
       
        
      </div>
    </div>
  );
}

export default App;
