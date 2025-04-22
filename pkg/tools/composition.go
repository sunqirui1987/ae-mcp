package tools

import (
	"encoding/json"
	"fmt"
	"github.com/sunqirui1987/ae-mcp/pkg/ae"
)

// CompositionDetails represents the details of an After Effects composition
type CompositionDetails map[string]interface{}

// GetCompositionDetails retrieves detailed information about a composition
func GetCompositionDetails(compositionName string) (CompositionDetails, error) {
	// Execute JavaScript to get composition details
	script := `
	try {
		var compName = "` + compositionName + `";
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
		var compDetails CompositionDetails
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
func CreateComposition(name string, width int, height int, duration float64, frameRate float64) (CompositionDetails, error) {
	// Execute JavaScript to create composition
	script := `
	try {
		var name = "` + name + `";
		var width = ` + fmt.Sprintf("%d", width) + `;
		var height = ` + fmt.Sprintf("%d", height) + `;
		var duration = ` + fmt.Sprintf("%f", duration) + `;
		var frameRate = ` + fmt.Sprintf("%f", frameRate) + `;
		
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
		var compDetails CompositionDetails
		if err := json.Unmarshal([]byte(resultStr), &compDetails); err != nil {
			return nil, err
		}
		
		return compDetails, nil
	}

	return nil, ErrInvalidResponse
} 