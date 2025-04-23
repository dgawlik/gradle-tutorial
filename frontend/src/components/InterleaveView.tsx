import React, { useState } from 'react';

interface WordHoverPopup {
  word: string;
  meanings: string[];
  position: { x: number; y: number };
}

interface InterleaveViewProps {
  originalText: string;
  translatedText: string;
  translationId: number;
}

const InterleaveView: React.FC<InterleaveViewProps> = ({ originalText, translatedText, translationId }) => {
  const [hoveredWord, setHoveredWord] = useState<string | null>(null);
  const [wordDefinitions, setWordDefinitions] = useState<string[]>([]);
  const [popupPosition, setPopupPosition] = useState<{ x: number, y: number }>({ x: 0, y: 0 });

  // Split text into sentences (by dot, exclamation mark, question mark, etc.)
  const splitIntoSentences = (text: string): string[] => {
    // Regular expression to split by sentence endings but keep the punctuation
    return text.split(/(?<=[.!?])\s+/).filter(Boolean);
  };

  const originalSentences = splitIntoSentences(originalText);
  const translatedSentences = splitIntoSentences(translatedText);

  // Handle word hover to show definitions
  const handleWordHover = async (word: string, event: React.MouseEvent) => {
    setHoveredWord(word);
    setPopupPosition({ 
      x: event.clientX, 
      y: event.clientY + window.scrollY + 20 // Position below the word
    });
    
    try {

      const meanings: string[] = await fetch('/api/definitions/'+ word).then(response => response.json());

      setWordDefinitions(meanings);
    } catch (error) {
      console.error('Failed to get word definitions:', error);
      setWordDefinitions([]);
    }
  };

  // Render a sentence with hoverable words
  const renderSentenceWithHover = (sentence: string, isOriginal: boolean) => {
    return (
      <div className={`p-2 ${isOriginal ? 'bg-gray-50' : 'bg-white'} rounded mb-1`}>
        {sentence.split(' ').map((word, idx) => {
          // Remove punctuation for lookup but keep it for display
          const cleanWord = word.replace(/[.,!?;:""]/g, '');
          return (
            <span
              key={idx}
              className={`${isOriginal ? 'hover:bg-blue-100 cursor-pointer' : ''} p-1 rounded transition-colors`}
              onMouseEnter={isOriginal ? (e) => handleWordHover(cleanWord, e) : undefined}
              onMouseLeave={isOriginal ? () => setHoveredWord(null) : undefined}
            >
              {word}{' '}
            </span>
          );
        })}
      </div>
    );
  };

  // Count valid sentences (skip empty ones if any)
  const sentenceCount = Math.min(
    originalSentences.filter(s => s.trim()).length,
    translatedSentences.filter(s => s.trim()).length
  );

  return (
    <div className="w-full">
      {/* Display interleaved sentences */}
      {Array.from({ length: sentenceCount }).map((_, i) => (
        <div key={i} className="mb-4 border-l-4 border-gray-300 pl-3">
          {renderSentenceWithHover(originalSentences[i], true)}
          {renderSentenceWithHover(translatedSentences[i], false)}
        </div>
      ))}

      {/* Word definition popup */}
      {hoveredWord && wordDefinitions && wordDefinitions.length > 0 && (
        <div 
          className="fixed bg-white shadow-lg border border-gray-200 rounded p-3 z-50 max-w-xs"
          style={{ 
            top: `${popupPosition.y}px`, 
            left: `${popupPosition.x}px`,
            position: 'absolute',
            transform: 'translateX(-50%)'
          }}
        >
          <p className="font-semibold text-gray-800 border-b pb-1 mb-2">{hoveredWord}:</p>
          <ul className="list-disc pl-5 text-gray-700 text-sm">
            {wordDefinitions.map((meaning, idx) => (
              <li key={idx} className="mb-1">{meaning}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default InterleaveView;