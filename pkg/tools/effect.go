package tools

import (
	"encoding/json"
	"fmt"
	"github.com/sunqirui1987/ae-mcp/pkg/ae"
)

// EffectDetails represents details of an After Effects effect
type EffectDetails map[string]interface{}

// EffectParameters represents parameters for an After Effects effect
type EffectParameters map[string]interface{}



// ApplyEffect applies an effect to a layer in a composition
func ApplyEffect(compName string, layerName string, effectName string, parameters EffectParameters) (EffectDetails, error) {

	
	// Build the script to apply the effect
	script := `
	try {
		var compName = "` + compName + `";
		var layerName = "` + layerName + `";
		var effectName = "` + effectName + `";
		
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
		
		// Find the layer
		var layer = null;
		for (var i = 1; i <= comp.numLayers; i++) {
			var l = comp.layer(i);
			if (l.name === layerName) {
				layer = l;
				break;
			}
		}
		
		if (!layer) {
			return JSON.stringify({
				error: "Layer not found: " + layerName
			});
		}
		
		// Apply the effect
		var effect = layer.Effects.addProperty(effectName);
		if (!effect) {
			return JSON.stringify({
				error: "Failed to apply effect: " + effectName
			});
		}
	`

	// If parameters were provided, adjust them
	if parameters != nil {
		for key, value := range parameters {
			// Different handling based on parameter type
			switch v := value.(type) {
			case float64:
				script += `
				try {
					if (effect.property("` + key + `")) {
						effect.property("` + key + `").setValue(` + fmt.Sprintf("%f", v) + `);
					}
				} catch (paramErr) {
					// Skip this parameter if it doesn't exist or can't be set
				}
				`
			case bool:
				boolVal := 0
				if v {
					boolVal = 1
				}
				script += `
				try {
					if (effect.property("` + key + `")) {
						effect.property("` + key + `").setValue(` + fmt.Sprintf("%d", boolVal) + `);
					}
				} catch (paramErr) {
					// Skip this parameter if it doesn't exist or can't be set
				}
				`
			case string:
				script += `
				try {
					if (effect.property("` + key + `")) {
						effect.property("` + key + `").setValue("` + escapeJSString(v) + `");
					}
				} catch (paramErr) {
					// Skip this parameter if it doesn't exist or can't be set
				}
				`
			case []interface{}:
				// Handle arrays (e.g., for color or point values)
				if len(v) > 0 {
					arrayStr := "["
					for i, item := range v {
						if i > 0 {
							arrayStr += ", "
						}
						switch itemVal := item.(type) {
						case float64:
							arrayStr += fmt.Sprintf("%f", itemVal)
						case int:
							arrayStr += fmt.Sprintf("%d", itemVal)
						case string:
							arrayStr += `"` + escapeJSString(itemVal) + `"`
						default:
							arrayStr += "0"
						}
					}
					arrayStr += "]"
					
					script += `
					try {
						if (effect.property("` + key + `")) {
							effect.property("` + key + `").setValue(` + arrayStr + `);
						}
					} catch (paramErr) {
						// Skip this parameter if it doesn't exist or can't be set
					}
					`
				}
			}
		}
	}

	// Complete the script
	script += `
		// Get information about the applied effect
		var effectInfo = {
			name: effect.name,
			matchName: effect.matchName,
			parameters: []
		};
		
		// Gather parameter information
		for (var i = 1; i <= effect.numProperties; i++) {
			var prop = effect.property(i);
			// Only include properties that we can potentially modify
			if (prop.canSetValue) {
				var paramInfo = {
					name: prop.name,
					matchName: prop.matchName,
					value: prop.value
				};
				effectInfo.parameters.push(paramInfo);
			}
		}
		
		return returnjson(effectInfo);
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
		var effectDetails EffectDetails
		if err := json.Unmarshal([]byte(resultStr), &effectDetails); err != nil {
			return nil, err
		}
		
		// Check for error in result
		if errMsg, hasErr := effectDetails["error"].(string); hasErr {
			return nil, fmt.Errorf("%s", errMsg)
		}
		
		return effectDetails, nil
	}

	return nil, ErrInvalidResponse
}
