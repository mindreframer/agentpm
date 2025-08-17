package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewXMLNormalizer(t *testing.T) {
	normalizer := NewXMLNormalizer()
	assert.NotNil(t, normalizer)

	// Should use default config
	expected := DefaultNormalizationConfig()
	assert.Equal(t, expected, normalizer.config)
}

func TestNewXMLNormalizerWithConfig(t *testing.T) {
	config := NormalizationConfig{
		IndentSize:          4,
		RemoveWhitespace:    false,
		SortAttributes:      false,
		NormalizeTimestamps: false,
		TimestampFields:     []string{"custom_field"},
	}

	normalizer := NewXMLNormalizerWithConfig(config)
	assert.NotNil(t, normalizer)
	assert.Equal(t, config, normalizer.config)
}

func TestDefaultNormalizationConfig(t *testing.T) {
	config := DefaultNormalizationConfig()

	assert.Equal(t, 2, config.IndentSize)
	assert.True(t, config.RemoveWhitespace)
	assert.True(t, config.SortAttributes)
	assert.True(t, config.NormalizeTimestamps)
	assert.Contains(t, config.TimestampFields, "created_at")
	assert.Contains(t, config.TimestampFields, "started_at")
	assert.Contains(t, config.TimestampFields, "completed_at")
}

func TestXMLNormalizer_EmptyInput(t *testing.T) {
	normalizer := NewXMLNormalizer()

	result, err := normalizer.NormalizeXML("")
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestXMLNormalizer_NonXMLInput(t *testing.T) {
	normalizer := NewXMLNormalizer()

	testCases := []string{
		"Plain text message",
		"Error: something went wrong",
		"No epic file specified",
		"Command completed successfully",
	}

	for _, input := range testCases {
		t.Run(input, func(t *testing.T) {
			result, err := normalizer.NormalizeXML(input)
			assert.NoError(t, err)
			assert.Equal(t, input, result)
		})
	}
}

func TestNormalizeXML_SimpleXML(t *testing.T) {
	normalizer := NewXMLNormalizer()

	input := `<epic id="test"><name>Test Epic</name></epic>`
	result, err := normalizer.NormalizeXML(input)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "<epic")
	assert.Contains(t, result, "<name>Test Epic</name>")
}

func TestXMLNormalizer_WithTimestamps(t *testing.T) {
	normalizer := NewXMLNormalizer()

	input := `<epic created_at="2025-08-16T09:00:00Z" status="active">
		<started_at>2025-08-16T10:00:00Z</started_at>
		<completed_at>2025-08-16T11:00:00Z</completed_at>
	</epic>`

	result, err := normalizer.NormalizeXML(input)

	assert.NoError(t, err)
	assert.Contains(t, result, `created_at="[TIMESTAMP]"`)
	assert.Contains(t, result, `<started_at>[TIMESTAMP]</started_at>`)
	assert.Contains(t, result, `<completed_at>[TIMESTAMP]</completed_at>`)
	assert.NotContains(t, result, "2025-08-16T09:00:00Z")
	assert.NotContains(t, result, "2025-08-16T10:00:00Z")
	assert.NotContains(t, result, "2025-08-16T11:00:00Z")
}

func TestXMLNormalizer_DisableTimestampNormalization(t *testing.T) {
	config := DefaultNormalizationConfig()
	config.NormalizeTimestamps = false
	normalizer := NewXMLNormalizerWithConfig(config)

	input := `<epic created_at="2025-08-16T09:00:00Z"><name>Test</name></epic>`
	result, err := normalizer.NormalizeXML(input)

	assert.NoError(t, err)
	assert.Contains(t, result, "2025-08-16T09:00:00Z")
	assert.NotContains(t, result, "[TIMESTAMP]")
}

func TestNormalizeXML_AttributeSorting(t *testing.T) {
	normalizer := NewXMLNormalizer()

	input := `<epic status="active" name="Test Epic" id="test-epic">`
	result, err := normalizer.NormalizeXML(input)

	assert.NoError(t, err)
	// Attributes should be sorted alphabetically
	assert.Contains(t, result, `id="test-epic" name="Test Epic" status="active"`)
}

func TestNormalizeXML_DisableAttributeSorting(t *testing.T) {
	config := DefaultNormalizationConfig()
	config.SortAttributes = false
	normalizer := NewXMLNormalizerWithConfig(config)

	input := `<epic status="active" id="test-epic">`
	result, err := normalizer.NormalizeXML(input)

	assert.NoError(t, err)
	// Attributes should remain in original order
	assert.Contains(t, result, `status="active" id="test-epic"`)
}

func TestNormalizeTimestamps(t *testing.T) {
	normalizer := NewXMLNormalizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Element with timestamp",
			input:    `<created_at>2025-08-16T09:00:00Z</created_at>`,
			expected: `<created_at>[TIMESTAMP]</created_at>`,
		},
		{
			name:     "Attribute with timestamp",
			input:    `<epic created_at="2025-08-16T09:00:00Z">`,
			expected: `<epic created_at="[TIMESTAMP]">`,
		},
		{
			name:     "Single quote attribute",
			input:    `<epic created_at='2025-08-16T09:00:00Z'>`,
			expected: `<epic created_at='[TIMESTAMP]'>`,
		},
		{
			name:     "Multiple timestamp fields",
			input:    `<epic created_at="2025-08-16T09:00:00Z"><started_at>2025-08-16T10:00:00Z</started_at></epic>`,
			expected: `<epic created_at="[TIMESTAMP]"><started_at>[TIMESTAMP]</started_at></epic>`,
		},
		{
			name:     "ISO8601 with milliseconds",
			input:    `<time>2025-08-16T09:00:00.123Z</time>`,
			expected: `<time>[TIMESTAMP]</time>`,
		},
		{
			name:     "ISO8601 with timezone",
			input:    `<time>2025-08-16T09:00:00+02:00</time>`,
			expected: `<time>[TIMESTAMP]</time>`,
		},
		{
			name:     "No timestamps",
			input:    `<epic id="test" name="Test Epic">content</epic>`,
			expected: `<epic id="test" name="Test Epic">content</epic>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizer.normalizeTimestamps(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSortAttributesInLine(t *testing.T) {
	normalizer := NewXMLNormalizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Two attributes",
			input:    `<epic status="active" id="test">`,
			expected: `<epic id="test" status="active">`,
		},
		{
			name:     "Self-closing tag",
			input:    `<test status="active" id="test"/>`,
			expected: `<test id="test" status="active"/>`,
		},
		{
			name:     "No attributes",
			input:    `<epic>content</epic>`,
			expected: `<epic>content</epic>`,
		},
		{
			name:     "Already sorted",
			input:    `<epic id="test" status="active">`,
			expected: `<epic id="test" status="active">`,
		},
		{
			name:     "Mixed quotes",
			input:    `<epic status="active" id='test'>`,
			expected: `<epic id='test' status="active">`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizer.sortAttributesInLine(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestXMLNormalizer_SortAttributes(t *testing.T) {
	normalizer := NewXMLNormalizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Two attributes",
			input:    `status="active" id="test"`,
			expected: `id="test" status="active"`,
		},
		{
			name:     "Three attributes",
			input:    `status="active" name="Test" id="test"`,
			expected: `id="test" name="Test" status="active"`,
		},
		{
			name:     "Single quotes",
			input:    `status='active' id='test'`,
			expected: `id='test' status='active'`,
		},
		{
			name:     "Mixed quotes",
			input:    `status="active" id='test'`,
			expected: `id='test' status="active"`,
		},
		{
			name:     "Already sorted",
			input:    `id="test" status="active"`,
			expected: `id="test" status="active"`,
		},
		{
			name:     "Empty input",
			input:    ``,
			expected: ``,
		},
		{
			name:     "Single attribute",
			input:    `id="test"`,
			expected: `id="test"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizer.sortAttributes(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRemoveExtraWhitespace(t *testing.T) {
	normalizer := NewXMLNormalizer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Multiple empty lines",
			input: `<epic>
			
			  <name>Test</name>
			  
			  
			</epic>`,
			expected: `<epic>
			  <name>Test</name>
			</epic>`,
		},
		{
			name: "Leading and trailing empty lines",
			input: `

			<epic><name>Test</name></epic>

			`,
			expected: `<epic><name>Test</name></epic>`,
		},
		{
			name:     "No empty lines",
			input:    `<epic><name>Test</name></epic>`,
			expected: `<epic><name>Test</name></epic>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizer.removeExtraWhitespace(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNormalizeXML_ComplexExample(t *testing.T) {
	normalizer := NewXMLNormalizer()

	// Complex XML with multiple features to normalize
	input := `<status epic="test-epic-1" format="xml" created_at="2025-08-16T09:00:00Z">
	
		<name>Test Epic</name>
		<status>active</status>
		
		<phases>
			<phase name="Setup" id="p1" started_at="2025-08-16T09:30:00Z" status="completed">
			
				<tasks>
					<task id="t1" status="completed" completed_at="2025-08-16T10:00:00Z"/>
				</tasks>
				
			</phase>
		</phases>
		
	</status>`

	result, err := normalizer.NormalizeXML(input)

	require.NoError(t, err)
	assert.NotEmpty(t, result)

	// Should normalize timestamps
	assert.Contains(t, result, `created_at="[TIMESTAMP]"`)
	assert.Contains(t, result, `started_at="[TIMESTAMP]"`)
	assert.Contains(t, result, `completed_at="[TIMESTAMP]"`)

	// Should sort attributes
	assert.Contains(t, result, `created_at="[TIMESTAMP]" epic="test-epic-1" format="xml"`)

	// Should preserve structure and content
	assert.Contains(t, result, "<name>Test Epic</name>")
	assert.Contains(t, result, "<status>active</status>")
}

// Benchmark tests for performance validation
func BenchmarkNormalizeXML_Simple(b *testing.B) {
	normalizer := NewXMLNormalizer()
	input := `<epic id="test" status="active"><name>Test Epic</name></epic>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := normalizer.NormalizeXML(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNormalizeXML_Complex(b *testing.B) {
	normalizer := NewXMLNormalizer()
	input := `<status epic="test-epic" created_at="2025-08-16T09:00:00Z" status="active">
		<name>Complex Epic</name>
		<phases>
			<phase id="p1" name="Setup" status="completed" started_at="2025-08-16T09:00:00Z" completed_at="2025-08-16T10:00:00Z">
				<tasks>
					<task id="t1" name="Task 1" status="completed"/>
					<task id="t2" name="Task 2" status="active" started_at="2025-08-16T10:30:00Z"/>
				</tasks>
			</phase>
		</phases>
	</status>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := normalizer.NormalizeXML(input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestNormalizeXML_ErrorHandling(t *testing.T) {
	normalizer := NewXMLNormalizer()

	// Test with malformed XML - should not error but return as-is
	malformedXML := `<epic><name>Test</epic>` // Missing closing tag
	result, err := normalizer.NormalizeXML(malformedXML)

	assert.NoError(t, err)
	assert.Equal(t, malformedXML, result)
}
