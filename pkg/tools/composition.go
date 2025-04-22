package tools

import (
	"encoding/json"
	"fmt"
	"github.com/sunqirui1987/ae-mcp/pkg/ae"
)

// GetCompositionDetails retrieves detailed information about a composition
func GetCompositionDetails(compositionName interface{}) (interface{}, error) {
	// Type check
	compName, ok := compositionName.(string)
	if !ok {
		return nil, ErrInvalidParams
	}

	// Execute JavaScript to get composition details
	script := `
	try {
		var compName = "` + compName + `";
		var project = app.project;
		var comp = null;
		
		// Find the composition by name
		for (var i = 1; i <= project.numItems; i++) {
			var item = project.item(i);
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
		
		// Get composition details
		var result = {
			name: comp.name,
			id: comp.id,
			duration: comp.duration,
			width: comp.width,
			height: comp.height,
			frameRate: comp.frameRate,
			frameDuration: comp.frameDuration,
			displayStartTime: comp.displayStartTime,
			workAreaStart: comp.workAreaStart,
			workAreaDuration: comp.workAreaDuration,
			numLayers: comp.numLayers,
			bgColor: [comp.bgColor[0], comp.bgColor[1], comp.bgColor[2]]
		};
		
		// Get layer information
		result.layers = [];
		for (var j = 1; j <= comp.numLayers; j++) {
			var layer = comp.layer(j);
			result.layers.push({
				name: layer.name,
				index: layer.index,
				enabled: layer.enabled,
				solo: layer.solo,
				shy: layer.shy,
				locked: layer.locked,
				hasVideo: layer.hasVideo,
				hasAudio: layer.hasAudio,
				is3D: layer.threeDLayer,
				position: layer.position.value,
				type: layer.source instanceof CompItem ? "Composition" :
					  layer.source instanceof FootageItem ? "Footage" :
					  layer.source instanceof SolidSource ? "Solid" :
					  "Unknown"
			});
		}
		
		return returnjson(result);
	} catch (err) {
		return "ERROR: " + err.toString();
	}
	`

	// Execute the script using the new file-based communication method
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
		var compDetails map[string]interface{}
		if err := json.Unmarshal([]byte(resultStr), &compDetails); err != nil {
			return nil, err
		}
		
		// Check for error in result
		if errMsg, hasErr := compDetails["error"].(string); hasErr {
			return nil, fmt.Errorf("%s", errMsg)
		}
		
		return compDetails, nil
	}

	return nil, ErrInvalidResponse
}

// CreateComposition creates a new composition in After Effects
func CreateComposition(name interface{}, width interface{}, height interface{}, duration interface{}, frameRate interface{}) (interface{}, error) {
	// Type checks
	nameStr, ok := name.(string)
	if !ok {
		return nil, fmt.Errorf("name must be a string: %w", ErrInvalidParams)
	}

	widthInt, ok := width.(int)
	if !ok {
		return nil, fmt.Errorf("width must be an integer: %w", ErrInvalidParams)
	}

	heightInt, ok := height.(int)
	if !ok {
		return nil, fmt.Errorf("height must be an integer: %w", ErrInvalidParams)
	}

	durationFloat, ok := duration.(float64)
	if !ok {
		return nil, fmt.Errorf("duration must be a number: %w", ErrInvalidParams)
	}

	frameRateFloat, ok := frameRate.(float64)
	if !ok {
		frameRateFloat = 30.0 // Default frame rate
	}

	// Execute JavaScript to create composition
	script := `
	try {
		var name = "` + nameStr + `";
		var width = ` + fmt.Sprintf("%d", widthInt) + `;
		var height = ` + fmt.Sprintf("%d", heightInt) + `;
		var duration = ` + fmt.Sprintf("%f", durationFloat) + `;
		var frameRate = ` + fmt.Sprintf("%f", frameRateFloat) + `;
		
		// Create the composition
		var project = app.project;
		var comp = project.items.addComp(name, width, height, 1, duration, frameRate);
		
		// Return the new composition details
		var result = {
			name: comp.name,
			id: comp.id,
			duration: comp.duration,
			width: comp.width,
			height: comp.height,
			frameRate: comp.frameRate
		};
		
		return returnjson(result);
	} catch (err) {
		return "ERROR: " + err.toString();
	}
	`

	// Execute the script using the new file-based communication method
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
		var compDetails map[string]interface{}
		if err := json.Unmarshal([]byte(resultStr), &compDetails); err != nil {
			return nil, err
		}
		
		return compDetails, nil
	}

	return nil, ErrInvalidResponse
} 