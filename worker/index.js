/**
 * 97LAYER OS Cloudflare Worker
 * Telegram webhook endpoint with zero-cost infrastructure
 * Handles incoming updates and forwards to processing
 */

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    // CORS headers for browser access
    const corsHeaders = {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
    };

    // Handle CORS preflight
    if (request.method === 'OPTIONS') {
      return new Response(null, { headers: corsHeaders });
    }

    // Health check endpoint
    if (url.pathname === '/health') {
      return new Response(JSON.stringify({
        status: 'healthy',
        service: '97LAYER OS Worker',
        timestamp: new Date().toISOString()
      }), {
        headers: {
          'Content-Type': 'application/json',
          ...corsHeaders
        }
      });
    }

    // Telegram webhook endpoint
    if (url.pathname === '/webhook' && request.method === 'POST') {
      try {
        const update = await request.json();

        // Validate Telegram update structure
        if (!update.update_id) {
          return new Response('Invalid update', { status: 400 });
        }

        // Extract message data
        const message = update.message || update.edited_message;
        if (!message) {
          return new Response('OK', { status: 200 });
        }

        const chatId = message.chat.id;
        const text = message.text || '';
        const photo = message.photo;

        // Log for debugging (visible in Cloudflare dashboard)
        console.log(`Received message from chat ${chatId}: ${text.slice(0, 100)}`);

        // Store in KV for processing (if configured)
        if (env.MESSAGES) {
          const messageKey = `msg_${update.update_id}`;
          await env.MESSAGES.put(messageKey, JSON.stringify({
            update_id: update.update_id,
            chat_id: chatId,
            text: text,
            has_photo: !!photo,
            timestamp: new Date().toISOString()
          }), {
            expirationTtl: 86400  // 24 hours
          });
        }

        // Forward to processing endpoint if configured
        if (env.PROCESSING_URL) {
          // Fire and forget - don't wait for response
          ctx.waitUntil(
            fetch(env.PROCESSING_URL, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'X-Worker-Auth': env.WORKER_AUTH_KEY || 'default'
              },
              body: JSON.stringify(update)
            }).catch(err => console.error('Forward failed:', err))
          );
        }

        // Always return 200 to Telegram quickly
        return new Response('OK', { status: 200 });

      } catch (error) {
        console.error('Webhook error:', error);
        return new Response('Internal error', { status: 500 });
      }
    }

    // Set webhook endpoint (one-time setup)
    if (url.pathname === '/setup' && url.searchParams.get('token')) {
      const token = url.searchParams.get('token');
      const webhookUrl = `${url.origin}/webhook`;

      try {
        // Set Telegram webhook
        const telegramUrl = `https://api.telegram.org/bot${token}/setWebhook`;
        const response = await fetch(telegramUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            url: webhookUrl,
            allowed_updates: ['message', 'edited_message', 'callback_query']
          })
        });

        const result = await response.json();

        if (result.ok) {
          return new Response(JSON.stringify({
            success: true,
            webhook_url: webhookUrl,
            message: 'Webhook set successfully'
          }), {
            headers: {
              'Content-Type': 'application/json',
              ...corsHeaders
            }
          });
        } else {
          return new Response(JSON.stringify({
            success: false,
            error: result.description
          }), {
            status: 400,
            headers: {
              'Content-Type': 'application/json',
              ...corsHeaders
            }
          });
        }
      } catch (error) {
        return new Response(JSON.stringify({
          success: false,
          error: error.message
        }), {
          status: 500,
          headers: {
            'Content-Type': 'application/json',
            ...corsHeaders
          }
        });
      }
    }

    // Stats endpoint (if KV configured)
    if (url.pathname === '/stats' && env.MESSAGES) {
      try {
        const list = await env.MESSAGES.list({ limit: 100 });
        const messageCount = list.keys.length;

        // Get recent messages
        const recentMessages = [];
        for (const key of list.keys.slice(0, 10)) {
          const data = await env.MESSAGES.get(key.name);
          if (data) {
            recentMessages.push(JSON.parse(data));
          }
        }

        return new Response(JSON.stringify({
          total_messages: messageCount,
          recent: recentMessages
        }), {
          headers: {
            'Content-Type': 'application/json',
            ...corsHeaders
          }
        });
      } catch (error) {
        return new Response(JSON.stringify({
          error: 'Stats unavailable'
        }), {
          status: 500,
          headers: {
            'Content-Type': 'application/json',
            ...corsHeaders
          }
        });
      }
    }

    // Default 404
    return new Response('Not Found', {
      status: 404,
      headers: corsHeaders
    });
  }
};