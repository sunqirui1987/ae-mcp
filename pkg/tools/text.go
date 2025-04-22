package tools

import (
	"encoding/json"
	"fmt"
	"github.com/sunqirui1987/ae-mcp/pkg/ae"
)

// TextLayerInfo represents information about a text layer
type TextLayerInfo map[string]interface{}

// TextOptions represents options for text layer creation or modification
type TextOptions struct {
	FontSize      float64   `json:"fontSize,omitempty"`
	FontName      string    `json:"fontName,omitempty"`
	Color         ColorRGB  `json:"color,omitempty"`
	Position      [2]float64 `json:"position,omitempty"`
	Justification string    `json:"justification,omitempty"`
}

// TextModifications represents modifications to be applied to a text layer
type TextModifications map[string]interface{}

// AddTextLayer adds a text layer to a composition
func AddTextLayer(compName string, text string, options *TextOptions) (TextLayerInfo, error) {
	// Default options
	fontSize := 72.0
	fontName := "Arial"
	color := [3]float64{1.0, 1.0, 1.0} // White
	position := [2]float64{0.0, 0.0}    // Center
	justification := "CENTER"
	
	// Apply provided options if available
	if options != nil {
		if options.FontSize > 0 {
			fontSize = options.FontSize
		}
		if options.FontName != "" {
			fontName = options.FontName
		}
		if options.Color != [3]float64{0, 0, 0} {
			color = options.Color
		}
		if options.Position != [2]float64{0, 0} {
			position = options.Position
		}
		if options.Justification != "" {
			justification = options.Justification
		}
	}

	// Construct the script to add a text layer
	script := `
	try {
		var compName = "` + compName + `";
		var textContent = "` + escapeJSString(text) + `";
		var fontSize = ` + fmt.Sprintf("%f", fontSize) + `;
		var fontName = "` + fontName + `";
		var color = [` + fmt.Sprintf("%f, %f, %f", color[0], color[1], color[2]) + `];
		var position = [` + fmt.Sprintf("%f, %f", position[0], position[1]) + `];
		var justification = ParagraphJustification.` + justification + `;
		
		// Find the composition
		var comp = null;
		for (var i = 1; i <= app.project.numItems; i++) {
			var item = app.project.item(i);
			if (item instanceof CompItem && item.name === compName) {
				comp = item;
				break;
			}
		}
		
		if (!comp) {
			return JSON.stringify({
				error: "Composition not found: " + compName
			});
		}
		
		// Add the text layer
		var textLayer = comp.layers.addText(textContent);
		var textProp = textLayer.property("Source Text");
		var textDocument = textProp.value;
		
		// Set text properties
		textDocument.fontSize = fontSize;
		textDocument.font = fontName;
		textDocument.fillColor = color;
		textDocument.justification = justification;
		
		// Apply the text document
		textProp.setValue(textDocument);
		
		// Set position if specified
		if (position[0] !== 0 || position[1] !== 0) {
			textLayer.position.setValue([position[0], position[1], 0]);
		} else {
			// Center the text in the composition
			textLayer.position.setValue([comp.width/2, comp.height/2, 0]);
		}
		
		// Return information about the created text layer
		var result = {
			name: textLayer.name,
			index: textLayer.index,
			text: textContent,
			fontSize: fontSize,
			fontName: fontName
		};
		
		return returnjson(result);
	} catch (err) {
		return "ERROR: " + err.toString();
	}
	`

	// Execute the script
	result, err := ae.ExecuteScript(script)
	if err != nil {
		return nil, err
	}

	// Extract result
	if resultStr, ok := result.(string); ok {
		// Check if the result indicates an error
		if len(resultStr) > 7 && resultStr[:7] == "ERROR: " {
			return nil, ErrAEScriptError(resultStr[7:])
		}
		
		// Parse the JSON result into a structured object
		var textLayerDetails TextLayerInfo
		if err := json.Unmarshal([]byte(resultStr), &textLayerDetails); err != nil {
			return nil, err
		}
		
		// Check for error in result
		if errMsg, hasErr := textLayerDetails["error"].(string); hasErr {
			return nil, fmt.Errorf("%s", errMsg)
		}
		
		return textLayerDetails, nil
	}

	return nil, ErrInvalidResponse
}

// ModifyTextLayer modifies an existing text layer in a composition
func ModifyTextLayer(compName string, layerName string, modifications TextModifications) (TextLayerInfo, error) {
	// Build the script
	script := `
	try {
		var compName = "` + compName + `";
		var layerName = "` + layerName + `";
		
		// Find the composition
		var comp = null;
		for (var i = 1; i <= app.project.numItems; i++) {
			var item = app.project.item(i);
			if (item instanceof CompItem && item.name === compName) {
				comp = item;
				break;
			}
		}
		
		if (!comp) {
			return JSON.stringify({
				error: "Composition not found: " + compName
			});
		}
		
		// Find the text layer
		var textLayer = null;
		for (var i = 1; i <= comp.numLayers; i++) {
			var layer = comp.layer(i);
			if (layer.name === layerName) {
				textLayer = layer;
				break;
			}
		}
		
		if (!textLayer) {
			return JSON.stringify({
				error: "Layer not found: " + layerName
			});
		}
		
		// Check if this is actually a text layer
		if (!(textLayer.property("Source Text") instanceof TextProperty)) {
			return JSON.stringify({
				error: "Layer is not a text layer: " + layerName
			});
		}
		
		// Get the text document
		var textProp = textLayer.property("Source Text");
		var textDocument = textProp.value;
		
		// Apply modifications
		var modified = false;
	`

	// Add modifications to the script
	if text, ok := modifications["text"].(string); ok {
		script += `
		textDocument.text = "` + escapeJSString(text) + `";
		modified = true;
		`
	}

	if fontSize, ok := modifications["fontSize"].(float64); ok {
		script += `
		textDocument.fontSize = ` + fmt.Sprintf("%f", fontSize) + `;
		modified = true;
		`
	}

	if fontName, ok := modifications["fontName"].(string); ok {
		script += `
		textDocument.font = "` + fontName + `";
		modified = true;
		`
	}

	if color, ok := modifications["color"].([]interface{}); ok && len(color) >= 3 {
		r, _ := color[0].(float64)
		g, _ := color[1].(float64)
		b, _ := color[2].(float64)
		script += `
		textDocument.fillColor = [` + fmt.Sprintf("%f, %f, %f", r, g, b) + `];
		modified = true;
		`
	}

	if justification, ok := modifications["justification"].(string); ok {
		script += `
		textDocument.justification = ParagraphJustification.` + justification + `;
		modified = true;
		`
	}

	if position, ok := modifications["position"].([]interface{}); ok && len(position) >= 2 {
		x, _ := position[0].(float64)
		y, _ := position[1].(float64)
		script += `
		textLayer.position.setValue([` + fmt.Sprintf("%f, %f, 0", x, y) + `]);
		modified = true;
		`
	}

	// Complete the script
	script += `
		// Apply the text document if modified
		if (modified) {
			textProp.setValue(textDocument);
		}
		
		// Return information about the modified text layer
		var result = {
			name: textLayer.name,
			index: textLayer.index,
			text: textDocument.text,
			fontSize: textDocument.fontSize,
			fontName: textDocument.font
		};
		
		return returnjson(result);
	} catch (err) {
		return "ERROR: " + err.toString();
	}
	`

	// Execute the script
	result, err := ae.ExecuteScript(script)
	if err != nil {
		return nil, err
	}

	// Extract result
	if resultStr, ok := result.(string); ok {
		// Check if the result indicates an error
		if len(resultStr) > 7 && resultStr[:7] == "ERROR: " {
			return nil, ErrAEScriptError(resultStr[7:])
		}
		
		// Parse the JSON result into a structured object
		var textLayerDetails TextLayerInfo
		if err := json.Unmarshal([]byte(resultStr), &textLayerDetails); err != nil {
			return nil, err
		}
		
		// Check for error in result
		if errMsg, hasErr := textLayerDetails["error"].(string); hasErr {
			return nil, fmt.Errorf("%s", errMsg)
		}
		
		return textLayerDetails, nil
	}

	return nil, ErrInvalidResponse
}

// Helper function to escape special characters in JS strings
func escapeJSString(s string) string {
	// Implement string escaping for JavaScript
	// This is a simple implementation; you might want to make it more robust
	result := ""
	for _, c := range s {
		switch c {
		case '\\':
			result += "\\\\"
		case '"':
			result += "\\\""
		case '\'':
			result += "\\'"
		case '\n':
			result += "\\n"
		case '\r':
			result += "\\r"
		case '\t':
			result += "\\t"
		default:
			result += string(c)
		}
	}
	return result
} 