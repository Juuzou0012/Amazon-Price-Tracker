import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

const String kApiBase = 'http://localhost:8080';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Amazon Price Tracker',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: const PriceHomePage(),
    );
  }
}

class PriceHomePage extends StatefulWidget {
  const PriceHomePage({super.key});
  @override
  State<PriceHomePage> createState() => _PriceHomePageState();
}

class _PriceHomePageState extends State<PriceHomePage> {
  final TextEditingController _inputController = TextEditingController();
  PriceData? _data;
  String? _error;
  bool _loading = false;

  Future<void> _fetchPrice() async {
    final input = _inputController.text.trim();
    if (input.isEmpty) return;

    // Pick asin in URL
    final match = RegExp(r'/([A-Z0-9]{10})(?:[/?]|$)').firstMatch(input);
    final asin = match != null ? match.group(1)! : input;

    setState(() {
      _loading = true;
      _error = null;
      _data = null;
    });

    try {
      final uri = Uri.parse('$kApiBase/api/price')
          .replace(queryParameters: {'asin': asin});
      final resp = await http.get(uri).timeout(const Duration(seconds: 5));
      if (resp.statusCode != 200) {
        throw Exception('Status ${resp.statusCode}');
      }
      final jsonMap = json.decode(resp.body) as Map<String, dynamic>;
      setState(() => _data = PriceData.fromJson(jsonMap));
    } catch (e) {
      setState(() {
        _error = 'Error fetching: $e';
      });
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Price Tracker')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            TextField(
              controller: _inputController,
              decoration: const InputDecoration(
                labelText: 'ASIN or link Amazon',
                hintText: 'Paste the full URL or just the ASIN',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            ElevatedButton.icon(
              onPressed: _loading ? null : _fetchPrice,
              icon: _loading
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.search),
              label: const Text('Search Price'),
            ),
            const SizedBox(height: 24),
            if (_error != null)
              Center(
                child: Text(
                  _error!,
                  style: const TextStyle(color: Colors.red),
                ),
              )
            else if (_data != null)
              Expanded(
                child: ListView(
                  children: [
                    Card(
                      elevation: 4,
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'ASIN: ${_data!.asin}',
                              style:
                                  const TextStyle(fontWeight: FontWeight.bold),
                            ),
                            const SizedBox(height: 8),
                            Text(
                                'Maior: \$${_data!.highest.toStringAsFixed(2)}'),
                            Text('Menor: \$${_data!.lowest.toStringAsFixed(2)}'),
                            Text(
                                'Atual: \$${_data!.current.toStringAsFixed(2)}'),
                            Text(
                                'Média: \$${_data!.average.toStringAsFixed(2)}'),
                            const SizedBox(height: 12),

                            // Full Image
                            Image.network(
                              '$kApiBase/api/image?asin=${_data!.asin}',
                              fit: BoxFit.contain,
                              loadingBuilder: (ctx, child, progress) =>
                                  progress == null
                                      ? child
                                      : const Center(
                                          child:
                                              CircularProgressIndicator()),
                            ),

                            const SizedBox(height: 12),
                            Text(
                              'Fetched at: ${_data!.fetchedAt.toLocal()}',
                              style: const TextStyle(
                                  fontSize: 12, color: Colors.grey),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              )
            else
              const Spacer(),
          ],
        ),
      ),
    );
  }
}

class PriceData {
  final String asin;
  final double highest;
  final double lowest;
  final double current;
  final double average;
  final DateTime fetchedAt;

  PriceData({
    required this.asin,
    required this.highest,
    required this.lowest,
    required this.current,
    required this.average,
    required this.fetchedAt,
  });

  factory PriceData.fromJson(Map<String, dynamic> json) {
    return PriceData(
      asin: json['asin'] as String,
      highest: (json['highest'] as num).toDouble(),
      lowest: (json['lowest'] as num).toDouble(),
      current: (json['current'] as num).toDouble(),
      average: (json['average'] as num).toDouble(),
      fetchedAt: DateTime.parse(json['fetched_at'] as String),
    );
  }
}
