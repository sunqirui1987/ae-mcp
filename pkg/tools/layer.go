package tools

import (
	"encoding/json"
	"fmt"
	"github.com/sunqirui1987/ae-mcp/pkg/ae"
)

// AddSolidLayer adds a solid layer to a composition
func AddSolidLayer(compositionName interface{}, layerName interface{}, color interface{}, width interface{}, height interface{}, is3D interface{}) (interface{}, error) {
	// Type checks
	compName, ok := compositionName.(string)
	if !ok {
		return nil, fmt.Errorf("composition_name must be a string: %w", ErrInvalidParams)
	}

	layerNameStr, ok := layerName.(string)
	if !ok {
		return nil, fmt.Errorf("layer_name must be a string: %w", ErrInvalidParams)
	}

	colorArray, ok := color.([]interface{})
	if !ok || len(colorArray) != 3 {
		return nil, fmt.Errorf("color must be an array of three values [R, G, B]: %w", ErrInvalidParams)
	}

	// Parse color values
	var r, g, b float64
	r, ok = colorArray[0].(float64)
	if !ok || r < 0 || r > 1 {
		return nil, fmt.Errorf("red color value must be between 0 and 1: %w", ErrInvalidParams)
	}
	g, ok = colorArray[1].(float64)
	if !ok || g < 0 || g > 1 {
		return nil, fmt.Errorf("green color value must be between 0 and 1: %w", ErrInvalidParams)
	}
	b, ok = colorArray[2].(float64)
	if !ok || b < 0 || b > 1 {
		return nil, fmt.Errorf("blue color value must be between 0 and 1: %w", ErrInvalidParams)
	}

	widthInt, _ := width.(int)
	heightInt, _ := height.(int)
	is3DBool, _ := is3D.(bool)

	// Execute JavaScript to add solid layer
	script := `
	try {
		var compName = "` + compName + `";
		var layerName = "` + layerNameStr + `";
		var color = [` + fmt.Sprintf("%f, %f, %f", r, g, b) + `];
		var width = ` + fmt.Sprintf("%d", widthInt) + `;
		var height = ` + fmt.Sprintf("%d", heightInt) + `;
		var is3D = ` + fmt.Sprintf("%t", is3DBool) + `;
		
		// Find the composition
		var project = app.project;
		var comp = null;
		
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
		
		// Use composition dimensions if width/height are not specified
		if (width <= 0) width = comp.width;
		if (height <= 0) height = comp.height;
		
		// Create solid
		var solidItem = project.items.addSolid(
			color,               // color array [r, g, b]
			layerName,           // name
			width,               // width
			height,              // height
			1,                   // pixel aspect ratio
			comp.duration        // duration (seconds)
		);
		
		// Add solid to comp
		var layer = comp.layers.add(solidItem);
		
		// Set 3D if specified
		if (is3D) {
			layer.threeDLayer = true;
		}
		
		// Return layer info
		var result = {
			name: layer.name,
			index: layer.index,
			enabled: layer.enabled,
			is3D: layer.threeDLayer,
			position: layer.position.value
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
		var layerInfo map[string]interface{}
		if err := json.Unmarshal([]byte(resultStr), &layerInfo); err != nil {
			return nil, err
		}
		
		// Check for error in result
		if errMsg, hasErr := layerInfo["error"].(string); hasErr {
			return nil, fmt.Errorf("%s", errMsg)
		}
		
		return layerInfo, nil
	}

	return nil, ErrInvalidResponse
}

// ModifyLayer modifies properties of an existing layer
func ModifyLayer(compositionName interface{}, layerIdentifier interface{}, properties interface{}) (interface{}, error) {
	// Type checks
	compName, ok := compositionName.(string)
	if !ok {
		return nil, fmt.Errorf("composition_name must be a string: %w", ErrInvalidParams)
	}

	layerID, ok := layerIdentifier.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("layer_identifier must be an object with name or index: %w", ErrInvalidParams)
	}

	props, ok := properties.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("properties must be an object: %w", ErrInvalidParams)
	}

	// Convert properties to JavaScript
	propsJSON, err := json.Marshal(props)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize properties: %w", err)
	}

	// Create layer identification JavaScript code
	var layerIdentifierJS string
	if layerName, hasName := layerID["name"].(string); hasName {
		layerIdentifierJS = `
		// Find layer by name
		var targetLayer = null;
		for (var i = 1; i <= comp.numLayers; i++) {
			if (comp.layer(i).name === "` + layerName + `") {
				targetLayer = comp.layer(i);
				break;
			}
		}
		`
	} else if layerIndex, hasIndex := layerID["index"].(float64); hasIndex {
		layerIdentifierJS = fmt.Sprintf(`
		// Get layer by index
		var targetLayer = comp.layer(%d);
		`, int(layerIndex))
	} else {
		return nil, fmt.Errorf("layer_identifier must have either name or index field: %w", ErrInvalidParams)
	}

	// Execute JavaScript to modify layer
	script := `
	try {
		var compName = "` + compName + `";
		var props = ` + string(propsJSON) + `;
		
		// Find the composition
		var project = app.project;
		var comp = null;
		
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
		
		` + layerIdentifierJS + `
		
		if (!targetLayer) {
			return JSON.stringify({
				error: "Layer not found"
			});
		}
		
		// Apply properties
		var result = {
			name: targetLayer.name,
			index: targetLayer.index,
			modified: {}
		};
		
		// Position
		if (props.position) {
			if (targetLayer.position.dimensionsSeparated) {
				// If dimensions are separated, we need to set X, Y, Z separately
				if (props.position[0] !== undefined) {
					targetLayer.transform.xPosition.setValue(props.position[0]);
					result.modified.xPosition = props.position[0];
				}
				if (props.position[1] !== undefined) {
					targetLayer.transform.yPosition.setValue(props.position[1]);
					result.modified.yPosition = props.position[1];
				}
				if (props.position[2] !== undefined && targetLayer.threeDLayer) {
					targetLayer.transform.zPosition.setValue(props.position[2]);
					result.modified.zPosition = props.position[2];
				}
			} else {
				// Set position as array
				targetLayer.position.setValue(props.position);
				result.modified.position = props.position;
			}
		}
		
		// Scale
		if (props.scale) {
			targetLayer.scale.setValue(props.scale);
			result.modified.scale = props.scale;
		}
		
		// Rotation
		if (props.rotation !== undefined) {
			targetLayer.rotation.setValue(props.rotation);
			result.modified.rotation = props.rotation;
		}
		
		// Opacity
		if (props.opacity !== undefined) {
			targetLayer.opacity.setValue(props.opacity);
			result.modified.opacity = props.opacity;
		}
		
		// Enabled
		if (props.enabled !== undefined) {
			targetLayer.enabled = props.enabled;
			result.modified.enabled = props.enabled;
		}
		
		// 3D Layer
		if (props.threeDLayer !== undefined) {
			targetLayer.threeDLayer = props.threeDLayer;
			result.modified.threeDLayer = props.threeDLayer;
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
		var resultData map[string]interface{}
		if err := json.Unmarshal([]byte(resultStr), &resultData); err != nil {
			return nil, err
		}
		
		// Check for error in result
		if errMsg, hasErr := resultData["error"].(string); hasErr {
			return nil, fmt.Errorf("%s", errMsg)
		}
		
		return resultData, nil
	}

	return nil, ErrInvalidResponse
} 