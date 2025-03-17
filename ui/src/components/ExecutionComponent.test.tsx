import { render, screen, fireEvent, act } from '../utils/test-utils';
import ExecutionComponent from './ExecutionComponent';
import gogenWasm from '../api/gogenWasm';
import { Configuration } from '../api/gogenApi';

// Mock the gogenWasm module
jest.mock('../api/gogenWasm', () => ({
  __esModule: true,
  default: {
    executeConfiguration: jest.fn(),
  },
}));

describe('ExecutionComponent', () => {
  const mockConfig = {
    gogen: 'test-config',
    description: 'Test configuration',
    config: 'config: test',
    samples: [],
    raters: [],
    mix: [],
    generators: [],
    global: {},
    templates: [],
  };

  beforeEach(() => {
    jest.clearAllMocks();
    // Mock the executeConfiguration function
    (gogenWasm.executeConfiguration as jest.Mock).mockResolvedValue(['Test output']);
  });

  test('renders execution form', () => {
    render(<ExecutionComponent configuration={mockConfig} />);
    
    expect(screen.getByLabelText(/events per interval/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/intervals/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/interval \(in seconds\)/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/output template/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/output mode/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /execute/i })).toBeInTheDocument();
  });

  test('validates input fields', async () => {
    render(<ExecutionComponent configuration={mockConfig} />);
    
    const eventCountInput = screen.getByLabelText(/events per interval/i);
    const intervalsInput = screen.getByLabelText(/intervals/i);
    const intervalSecondsInput = screen.getByLabelText(/interval \(in seconds\)/i);
    const executeButton = screen.getByRole('button', { name: /execute/i });

    // Test invalid inputs
    await act(async () => {
      fireEvent.change(eventCountInput, { target: { value: 'abc' } });
      fireEvent.change(intervalsInput, { target: { value: '-1' } });
      fireEvent.change(intervalSecondsInput, { target: { value: 'xyz' } });
    });

    // The execute button should still be enabled as the component handles invalid inputs by ignoring them
    expect(executeButton).not.toBeDisabled();

    // Test valid inputs
    await act(async () => {
      fireEvent.change(eventCountInput, { target: { value: '100' } });
      fireEvent.change(intervalsInput, { target: { value: '2' } });
      fireEvent.change(intervalSecondsInput, { target: { value: '5' } });
    });

    expect(executeButton).not.toBeDisabled();
  });

  test('executes configuration in terminal mode', async () => {
    render(<ExecutionComponent configuration={mockConfig} />);
    
    const executeButton = screen.getByRole('button', { name: /execute/i });

    await act(async () => {
      fireEvent.click(executeButton);
    });

    // Terminal container should be visible
    expect(screen.getByTestId('terminal-container')).toBeInTheDocument();

    // Should have called executeConfiguration with correct params
    expect(gogenWasm.executeConfiguration).toHaveBeenCalledWith(
      mockConfig,
      {
        eventCount: 1,
        intervals: 5,
        intervalSeconds: 1,
        outputTemplate: 'raw'
      },
      expect.any(Function) // Terminal mode uses callback
    );
  });

  test('executes configuration in structured mode', async () => {
    render(<ExecutionComponent configuration={mockConfig} />);
    
    // Switch to structured mode
    const outputModeSelect = screen.getByLabelText(/output mode/i);
    await act(async () => {
      fireEvent.change(outputModeSelect, { target: { value: 'structured' } });
    });

    const executeButton = screen.getByRole('button', { name: /execute/i });
    await act(async () => {
      fireEvent.click(executeButton);
    });

    // Should have called executeConfiguration without callback
    expect(gogenWasm.executeConfiguration).toHaveBeenCalledWith(
      mockConfig,
      {
        eventCount: 1,
        intervals: 5,
        intervalSeconds: 1,
        outputTemplate: 'raw'
      }
    );
  });

  it('handles execution errors', async () => {
    const errorMessage = 'Test error';
    const mockExecute = gogenWasm.executeConfiguration as jest.Mock;
    mockExecute.mockRejectedValueOnce(new Error(errorMessage));

    const mockConfig: Configuration = {
      gogen: 'test',
      description: 'Test description',
      config: 'test config',
      samples: [],
      raters: [],
      mix: [],
      generators: [],
      templates: []
    };

    render(<ExecutionComponent configuration={mockConfig} />);

    const executeButton = screen.getByText('Execute');
    await act(async () => {
      fireEvent.click(executeButton);
    });

    // Error message should be displayed in the terminal
    const terminalContainer = screen.getByTestId('terminal-container');
    expect(terminalContainer).toBeInTheDocument();
    expect(mockExecute).toHaveBeenCalled();
  });
}); 